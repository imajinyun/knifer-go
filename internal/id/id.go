package id

import (
	cryptorand "crypto/rand"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"hash/fnv"
	mathrand "math/rand"
	"net"
	"os"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

const (
	nanoIDAlphabet = "_-0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

	snowflakeWorkerIDBits     = int64(5)
	snowflakeDatacenterIDBits = int64(5)
	snowflakeSequenceBits     = int64(12)
	snowflakeMaxWorkerID      = int64(-1) ^ (int64(-1) << snowflakeWorkerIDBits)
	snowflakeMaxDatacenterID  = int64(-1) ^ (int64(-1) << snowflakeDatacenterIDBits)
	snowflakeSequenceMask     = int64(-1) ^ (int64(-1) << snowflakeSequenceBits)
	snowflakeWorkerIDShift    = snowflakeSequenceBits
	snowflakeDatacenterShift  = snowflakeSequenceBits + snowflakeWorkerIDBits
	snowflakeTimestampShift   = snowflakeSequenceBits + snowflakeWorkerIDBits + snowflakeDatacenterIDBits
	snowflakeEpoch            = int64(1288834974657)
)

var (
	defaultRand      = mathrand.New(mathrand.NewSource(time.Now().UnixNano())) // #nosec G404 -- fallback only for IDs when crypto/rand is unavailable.
	defaultRandMu    sync.Mutex
	objectIDCounter  uint32
	snowflakeCache   sync.Map
	defaultSnowflake atomic.Value
)

// RandomUUID returns a standard random UUID string in 8-4-4-4-12 format.
func RandomUUID() string { return formatUUID(randomUUIDBytes(), false) }

// SimpleUUID returns a 32-character UUID without hyphens.
func SimpleUUID() string { return formatUUID(randomUUIDBytes(), true) }

// FastUUID returns a standard random UUID string.
// Go uses crypto/rand directly here; the name is kept as a convenient alias.
func FastUUID() string { return RandomUUID() }

// FastSimpleUUID returns a 32-character UUID without hyphens.
// Go uses crypto/rand directly here; the name is kept as a convenient alias.
func FastSimpleUUID() string { return SimpleUUID() }

func randomUUIDBytes() []byte {
	b := make([]byte, 16)
	fillRandomBytes(b)
	// version 4 / variant
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return b
}

func formatUUID(b []byte, simple bool) string {
	if simple {
		return hex.EncodeToString(b)
	}
	s := hex.EncodeToString(b)
	return s[0:8] + "-" + s[8:12] + "-" + s[12:16] + "-" + s[16:20] + "-" + s[20:]
}

// ObjectId returns a MongoDB-style 12-byte object id encoded as 24 hex characters.
// Layout: 4-byte Unix timestamp in seconds, 5 random bytes, and a 3-byte counter.
func ObjectId() string {
	now := uint32(time.Now().Unix()) // #nosec G115 -- ObjectId timestamp is intentionally stored in 4 bytes.
	rnd := make([]byte, 5)
	fillRandomBytes(rnd)
	c := nextCounter()
	b := make([]byte, 12)
	binary.BigEndian.PutUint32(b[0:4], now)
	copy(b[4:9], rnd)
	b[9] = byte(c >> 16)
	b[10] = byte(c >> 8)
	b[11] = byte(c)
	return hex.EncodeToString(b)
}

func nextCounter() uint32 { return atomic.AddUint32(&objectIDCounter, 1) & 0x00ffffff }

// Snowflake is a Twitter Snowflake-style ID generator.
// The generated int64 layout is: timestamp(41 bits), datacenter(5 bits), worker(5 bits), sequence(12 bits).
type Snowflake struct {
	mu            sync.Mutex
	workerID      int64
	datacenterID  int64
	sequence      int64
	lastTimestamp int64
	timeFunc      func() int64
	tilNextMillis func(lastTimestamp int64, now func() int64) int64
}

// CreateSnowflake creates a standalone Snowflake generator.
// Multiple standalone generators with the same worker/datacenter pair may produce duplicate IDs.
func CreateSnowflake(workerID, datacenterID int64) *Snowflake {
	return newSnowflake(workerID, datacenterID)
}

// GetSnowflake returns the default singleton Snowflake generator.
func GetSnowflake() *Snowflake {
	if v := defaultSnowflake.Load(); v != nil {
		return v.(*Snowflake)
	}
	dc := GetDataCenterID(snowflakeMaxDatacenterID)
	worker := GetWorkerID(dc, snowflakeMaxWorkerID)
	sf := GetSnowflakeWithWorkerDataCenter(worker, dc)
	defaultSnowflake.Store(sf)
	return sf
}

// GetSnowflakeWithWorker returns a singleton Snowflake generator for workerID.
func GetSnowflakeWithWorker(workerID int64) *Snowflake {
	return GetSnowflakeWithWorkerDataCenter(workerID, GetDataCenterID(snowflakeMaxDatacenterID))
}

// GetSnowflakeWithWorkerDataCenter returns a singleton Snowflake generator for worker/datacenter pair.
func GetSnowflakeWithWorkerDataCenter(workerID, datacenterID int64) *Snowflake {
	workerID = normalizeSnowflakeID(workerID, snowflakeMaxWorkerID)
	datacenterID = normalizeSnowflakeID(datacenterID, snowflakeMaxDatacenterID)
	key := fmt.Sprintf("%d:%d", workerID, datacenterID)
	if v, ok := snowflakeCache.Load(key); ok {
		return v.(*Snowflake)
	}
	sf := newSnowflake(workerID, datacenterID)
	actual, _ := snowflakeCache.LoadOrStore(key, sf)
	return actual.(*Snowflake)
}

func newSnowflake(workerID, datacenterID int64) *Snowflake {
	workerID = normalizeSnowflakeID(workerID, snowflakeMaxWorkerID)
	datacenterID = normalizeSnowflakeID(datacenterID, snowflakeMaxDatacenterID)
	return &Snowflake{
		workerID:      workerID,
		datacenterID:  datacenterID,
		lastTimestamp: -1,
		timeFunc:      currentMillis,
		tilNextMillis: waitNextMillis,
	}
}

// WorkerID returns the generator worker id.
func (s *Snowflake) WorkerID() int64 { return s.workerID }

// DatacenterID returns the generator datacenter id.
func (s *Snowflake) DatacenterID() int64 { return s.datacenterID }

// NextID returns the next Snowflake ID.
func (s *Snowflake) NextID() int64 {
	s.mu.Lock()
	defer s.mu.Unlock()

	timestamp := s.timeFunc()
	if timestamp < s.lastTimestamp {
		// Avoid returning non-monotonic IDs if system time moves backwards.
		timestamp = s.lastTimestamp
	}
	if timestamp == s.lastTimestamp {
		s.sequence = (s.sequence + 1) & snowflakeSequenceMask
		if s.sequence == 0 {
			timestamp = s.tilNextMillis(s.lastTimestamp, s.timeFunc)
		}
	} else {
		s.sequence = 0
	}
	s.lastTimestamp = timestamp
	return ((timestamp - snowflakeEpoch) << snowflakeTimestampShift) |
		(s.datacenterID << snowflakeDatacenterShift) |
		(s.workerID << snowflakeWorkerIDShift) |
		s.sequence
}

// NextIDStr returns the next Snowflake ID as a decimal string.
func (s *Snowflake) NextIDStr() string { return strconv.FormatInt(s.NextID(), 10) }

// GetDataCenterID derives a datacenter id from the local MAC address.
func GetDataCenterID(maxDatacenterID int64) int64 {
	if maxDatacenterID <= 0 {
		return 1
	}
	if maxDatacenterID == int64(^uint64(0)>>1) {
		maxDatacenterID--
	}
	for _, iface := range networkInterfaces() {
		mac := iface.HardwareAddr
		if len(mac) >= 2 {
			id := ((0x000000FF & int64(mac[len(mac)-2])) | (0x0000FF00 & (int64(mac[len(mac)-1]) << 8))) >> 6
			return id % (maxDatacenterID + 1)
		}
	}
	return 1 % (maxDatacenterID + 1)
}

// GetWorkerID derives a worker id from datacenter id and process id.
func GetWorkerID(datacenterID, maxWorkerID int64) int64 {
	if maxWorkerID <= 0 {
		return 0
	}
	h := fnv.New32a()
	_, _ = h.Write([]byte(strconv.FormatInt(datacenterID, 10)))
	_, _ = h.Write([]byte(strconv.Itoa(os.Getpid())))
	return int64(h.Sum32()&0xffff) % (maxWorkerID + 1)
}

// NanoId returns a default 21-character NanoId using a URL-safe alphabet.
func NanoId() string { return NanoIdN(21) }

// NanoIdN returns a NanoId with the specified length.
func NanoIdN(n int) string {
	if n <= 0 {
		return ""
	}
	mask := byte(63) // alphabet length is 64.
	step := (n*8 + 7) / 8
	out := make([]byte, 0, n)
	buf := make([]byte, step)
	for {
		fillRandomBytes(buf)
		for i := 0; i < step && len(out) < n; i++ {
			out = append(out, nanoIDAlphabet[buf[i]&mask])
		}
		if len(out) >= n {
			break
		}
	}
	return string(out[:n])
}

// GetSnowflakeNextID returns the next ID from the default singleton Snowflake generator.
func GetSnowflakeNextID() int64 { return GetSnowflake().NextID() }

// GetSnowflakeNextIDStr returns the next ID string from the default singleton Snowflake generator.
func GetSnowflakeNextIDStr() string { return GetSnowflake().NextIDStr() }

func fillRandomBytes(buf []byte) {
	if _, err := cryptorand.Read(buf); err != nil {
		defaultRandMu.Lock()
		defer defaultRandMu.Unlock()
		for i := range buf {
			buf[i] = byte(defaultRand.Intn(256)) // #nosec G115 -- Intn(256) always fits in byte.
		}
	}
}

func normalizeSnowflakeID(id, max int64) int64 {
	if max <= 0 {
		return 0
	}
	if id < 0 {
		id = -id
	}
	return id % (max + 1)
}

func currentMillis() int64 { return time.Now().UnixNano() / int64(time.Millisecond) }

func waitNextMillis(lastTimestamp int64, now func() int64) int64 {
	timestamp := now()
	for timestamp <= lastTimestamp {
		time.Sleep(time.Millisecond)
		timestamp = now()
	}
	return timestamp
}

func networkInterfaces() []net.Interface {
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil
	}
	return ifaces
}
