package id

import (
	cryptorand "crypto/rand"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"hash/fnv"
	"io"
	mathrand "math/rand"
	"net"
	"os"
	"strconv"
	"strings"
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
	defaultRand         *mathrand.Rand
	defaultRandMu       sync.Mutex
	defaultRandProvider = newDefaultFallbackRand
	objectIDCounter     uint32
	snowflakeCache      sync.Map
	defaultSnowflake    atomic.Value
)

type randomConfig struct {
	reader         io.Reader
	fallbackSource *mathrand.Rand
}

// RandomOption customizes random-byte based ID helpers.
type RandomOption func(*randomConfig)

// WithRandomReader sets the random source used by ID helpers.
func WithRandomReader(reader io.Reader) RandomOption {
	return func(c *randomConfig) {
		if reader != nil {
			c.reader = reader
		}
	}
}

// WithFallbackRandomSource sets the pseudo-random fallback used when the
// primary random reader fails. It exists for compatibility and deterministic
// tests; do not use it for security-sensitive identifiers.
func WithFallbackRandomSource(source *mathrand.Rand) RandomOption {
	return func(c *randomConfig) {
		if source != nil {
			c.fallbackSource = source
		}
	}
}

func applyRandomOptions(opts []RandomOption) randomConfig {
	cfg := randomConfig{reader: cryptorand.Reader}
	for _, opt := range opts {
		if opt != nil {
			opt(&cfg)
		}
	}
	if cfg.reader == nil {
		cfg.reader = cryptorand.Reader
	}
	return cfg
}

// ConfigureDefaultFallbackRandomSourceProvider sets the provider used to lazily
// create the package-level fallback PRNG when the primary random reader fails.
// The fallback PRNG is compatibility behavior for UUID/ObjectId helpers, not a
// cryptographic entropy source.
// Passing nil restores the time-seeded default provider and clears cached state.
func ConfigureDefaultFallbackRandomSourceProvider(provider func() *mathrand.Rand) {
	defaultRandMu.Lock()
	defer defaultRandMu.Unlock()
	defaultRand = nil
	if provider == nil {
		defaultRandProvider = newDefaultFallbackRand
		return
	}
	defaultRandProvider = provider
}

// ResetDefaultFallbackRandomSource restores the fallback PRNG provider and clears cached state.
func ResetDefaultFallbackRandomSource() { ConfigureDefaultFallbackRandomSourceProvider(nil) }

// SetFallbackRandomSeed resets the package-level fallback PRNG to a deterministic seed.
// It is intended for tests and reproducible non-security fallback behavior only.
func SetFallbackRandomSeed(seed int64) {
	defaultRandMu.Lock()
	defer defaultRandMu.Unlock()
	defaultRand = mathrand.New(mathrand.NewSource(seed))
}

type objectIDConfig struct {
	randomConfig
	now     func() time.Time
	counter func() uint32
}

// ObjectIDOption customizes ObjectIdWithOptions.
type ObjectIDOption func(*objectIDConfig)

// WithObjectIDRandomReader sets the random source used by ObjectIdWithOptions.
func WithObjectIDRandomReader(reader io.Reader) ObjectIDOption {
	return func(c *objectIDConfig) {
		if reader != nil {
			c.reader = reader
		}
	}
}

// WithObjectIDFallbackRandomSource sets the fallback pseudo-random source used
// when ObjectId random reads fail. It is intended for compatibility and tests,
// not for security-sensitive identifiers.
func WithObjectIDFallbackRandomSource(source *mathrand.Rand) ObjectIDOption {
	return func(c *objectIDConfig) {
		if source != nil {
			c.fallbackSource = source
		}
	}
}

// WithObjectIDTimeFunc sets the time source used by ObjectIdWithOptions.
func WithObjectIDTimeFunc(now func() time.Time) ObjectIDOption {
	return func(c *objectIDConfig) {
		if now != nil {
			c.now = now
		}
	}
}

// WithObjectIDCounter sets the counter source used by ObjectIdWithOptions.
func WithObjectIDCounter(counter func() uint32) ObjectIDOption {
	return func(c *objectIDConfig) {
		if counter != nil {
			c.counter = counter
		}
	}
}

func applyObjectIDOptions(opts []ObjectIDOption) objectIDConfig {
	cfg := objectIDConfig{randomConfig: randomConfig{reader: cryptorand.Reader}, now: time.Now, counter: nextCounter}
	for _, opt := range opts {
		if opt != nil {
			opt(&cfg)
		}
	}
	if cfg.reader == nil {
		cfg.reader = cryptorand.Reader
	}
	if cfg.now == nil {
		cfg.now = time.Now
	}
	if cfg.counter == nil {
		cfg.counter = nextCounter
	}
	return cfg
}

type nanoIDConfig struct {
	randomConfig
	alphabet string
	length   int
}

type snowflakeConfig struct {
	workerID      int64
	datacenterID  int64
	timeFunc      func() int64
	tilNextMillis func(lastTimestamp int64, now func() int64) int64
	interfaces    func() ([]net.Interface, error)
	pid           func() int
	workerSet     bool
	datacenterSet bool
	cache         bool
	cacheSet      bool
	runtimeSet    bool
}

// NanoIDOption customizes NanoIdWithOptions.
type NanoIDOption func(*nanoIDConfig)

// SnowflakeOption customizes Snowflake construction.
type SnowflakeOption func(*snowflakeConfig)

// WithNanoIDRandomReader sets the random source used by NanoIdWithOptions.
func WithNanoIDRandomReader(reader io.Reader) NanoIDOption {
	return func(c *nanoIDConfig) {
		if reader != nil {
			c.reader = reader
		}
	}
}

// WithNanoIDFallbackRandomSource sets the fallback random source used when NanoId random reads fail.
func WithNanoIDFallbackRandomSource(source *mathrand.Rand) NanoIDOption {
	return func(c *nanoIDConfig) {
		if source != nil {
			c.fallbackSource = source
		}
	}
}

// WithNanoIDAlphabet sets the alphabet used by NanoIdWithOptions.
func WithNanoIDAlphabet(alphabet string) NanoIDOption {
	return func(c *nanoIDConfig) { c.alphabet = alphabet }
}

// WithNanoIDLength sets the generated ID length used by NanoIdWithOptions.
func WithNanoIDLength(length int) NanoIDOption {
	return func(c *nanoIDConfig) { c.length = length }
}

func applyNanoIDOptions(opts []NanoIDOption) nanoIDConfig {
	cfg := nanoIDConfig{randomConfig: randomConfig{reader: cryptorand.Reader}, alphabet: nanoIDAlphabet, length: 21}
	for _, opt := range opts {
		if opt != nil {
			opt(&cfg)
		}
	}
	if cfg.reader == nil {
		cfg.reader = cryptorand.Reader
	}
	if cfg.alphabet == "" {
		cfg.alphabet = nanoIDAlphabet
	}
	return cfg
}

// WithSnowflakeWorkerID sets the generator worker id.
func WithSnowflakeWorkerID(workerID int64) SnowflakeOption {
	return func(c *snowflakeConfig) {
		c.workerID = workerID
		c.workerSet = true
	}
}

// WithSnowflakeDatacenterID sets the generator datacenter id.
func WithSnowflakeDatacenterID(datacenterID int64) SnowflakeOption {
	return func(c *snowflakeConfig) {
		c.datacenterID = datacenterID
		c.datacenterSet = true
	}
}

// WithSnowflakeTimeFunc sets the millisecond time source used by the generator.
func WithSnowflakeTimeFunc(timeFunc func() int64) SnowflakeOption {
	return func(c *snowflakeConfig) {
		if timeFunc != nil {
			c.timeFunc = timeFunc
			c.runtimeSet = true
		}
	}
}

// WithSnowflakeWaitFunc sets the wait function used when the sequence overflows within the same millisecond.
func WithSnowflakeWaitFunc(waitFunc func(lastTimestamp int64, now func() int64) int64) SnowflakeOption {
	return func(c *snowflakeConfig) {
		if waitFunc != nil {
			c.tilNextMillis = waitFunc
			c.runtimeSet = true
		}
	}
}

// WithSnowflakeCache controls whether singleton helper variants may reuse package-level cached generators.
func WithSnowflakeCache(enabled bool) SnowflakeOption {
	return func(c *snowflakeConfig) {
		c.cache = enabled
		c.cacheSet = true
	}
}

// WithSnowflakeInterfacesFunc sets the network interface provider used to derive the default datacenter id.
func WithSnowflakeInterfacesFunc(interfaces func() ([]net.Interface, error)) SnowflakeOption {
	return func(c *snowflakeConfig) {
		if interfaces != nil {
			c.interfaces = interfaces
		}
	}
}

// WithSnowflakePIDFunc sets the process-id provider used to derive the default worker id.
func WithSnowflakePIDFunc(pid func() int) SnowflakeOption {
	return func(c *snowflakeConfig) {
		if pid != nil {
			c.pid = pid
		}
	}
}

func applySnowflakeOptions(opts []SnowflakeOption) snowflakeConfig {
	cfg := snowflakeConfig{timeFunc: currentMillis, tilNextMillis: waitNextMillis, interfaces: net.Interfaces, pid: os.Getpid, cache: true}
	for _, opt := range opts {
		if opt != nil {
			opt(&cfg)
		}
	}
	if cfg.runtimeSet && !cfg.cacheSet {
		cfg.cache = false
	}
	if cfg.timeFunc == nil {
		cfg.timeFunc = currentMillis
	}
	if cfg.tilNextMillis == nil {
		cfg.tilNextMillis = waitNextMillis
	}
	if cfg.interfaces == nil {
		cfg.interfaces = net.Interfaces
	}
	if cfg.pid == nil {
		cfg.pid = os.Getpid
	}
	return cfg
}

func applyDefaultSnowflakeOptions(opts []SnowflakeOption) snowflakeConfig {
	cfg := applySnowflakeOptions(opts)
	if !cfg.datacenterSet {
		cfg.datacenterID = getDataCenterID(snowflakeMaxDatacenterID, cfg.interfaces)
	}
	if !cfg.workerSet {
		cfg.workerID = getWorkerID(cfg.datacenterID, snowflakeMaxWorkerID, cfg.pid)
	}
	return cfg
}

// RandomUUID returns a standard random UUID string in 8-4-4-4-12 format.
func RandomUUID() string { return RandomUUIDWithOptions() }

// RandomUUIDWithOptions returns a standard random UUID string using custom random options.
func RandomUUIDWithOptions(opts ...RandomOption) string {
	cfg := applyRandomOptions(opts)
	return formatUUID(randomUUIDBytesFrom(cfg), false)
}

// SimpleUUID returns a 32-character UUID without hyphens.
func SimpleUUID() string { return SimpleUUIDWithOptions() }

// SimpleUUIDWithOptions returns a 32-character UUID without hyphens using custom random options.
func SimpleUUIDWithOptions(opts ...RandomOption) string {
	cfg := applyRandomOptions(opts)
	return formatUUID(randomUUIDBytesFrom(cfg), true)
}

// FastUUID returns a standard random UUID string.
// Go uses crypto/rand directly here; the name is kept as a convenient alias.
func FastUUID() string { return RandomUUID() }

// FastUUIDWithOptions returns a standard random UUID string using custom random options.
func FastUUIDWithOptions(opts ...RandomOption) string { return RandomUUIDWithOptions(opts...) }

// FastSimpleUUID returns a 32-character UUID without hyphens.
// Go uses crypto/rand directly here; the name is kept as a convenient alias.
func FastSimpleUUID() string { return SimpleUUID() }

// FastSimpleUUIDWithOptions returns a 32-character UUID without hyphens using custom random options.
func FastSimpleUUIDWithOptions(opts ...RandomOption) string { return SimpleUUIDWithOptions(opts...) }

func randomUUIDBytesFrom(cfg randomConfig) []byte {
	b := make([]byte, 16)
	fillRandomBytesWithConfig(cfg, b)
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
	return ObjectIdWithOptions()
}

// ObjectIdWithOptions returns a MongoDB-style object id with custom generation options.
func ObjectIdWithOptions(opts ...ObjectIDOption) string {
	cfg := applyObjectIDOptions(opts)
	now := uint32(cfg.now().Unix()) // #nosec G115 -- ObjectId timestamp is intentionally stored in 4 bytes.
	rnd := make([]byte, 5)
	fillRandomBytesWithConfig(cfg.randomConfig, rnd)
	c := cfg.counter()
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
	return CreateSnowflakeWithOptions(WithSnowflakeWorkerID(workerID), WithSnowflakeDatacenterID(datacenterID))
}

// CreateSnowflakeWithOptions creates a standalone Snowflake generator customized by options.
func CreateSnowflakeWithOptions(opts ...SnowflakeOption) *Snowflake {
	cfg := applySnowflakeOptions(opts)
	return newSnowflakeWithConfig(cfg)
}

// NewIsolatedSnowflake creates a standalone Snowflake generator without singleton/cache lookup.
func NewIsolatedSnowflake(opts ...SnowflakeOption) *Snowflake {
	return newSnowflakeWithConfig(applyDefaultSnowflakeOptions(opts))
}

// GetSnowflake returns the default singleton Snowflake generator.
func GetSnowflake() *Snowflake {
	return GetSnowflakeWithOptions()
}

// GetSnowflakeWithOptions returns the default singleton Snowflake generator.
// If the default singleton has not been created yet, opts customize its worker,
// datacenter, clock, and wait strategy. Once created, later calls return the
// existing singleton; use ConfigureDefaultSnowflake to replace it deliberately.
func GetSnowflakeWithOptions(opts ...SnowflakeOption) *Snowflake {
	cfg := applyDefaultSnowflakeOptions(opts)
	if !cfg.cache {
		return newSnowflakeWithConfig(cfg)
	}
	if v := defaultSnowflake.Load(); v != nil {
		return v.(*Snowflake)
	}
	sf := newSnowflakeWithConfig(cfg)
	if defaultSnowflake.CompareAndSwap(nil, sf) {
		return sf
	}
	return defaultSnowflake.Load().(*Snowflake)
}

// ConfigureDefaultSnowflake replaces the default singleton Snowflake generator.
// It is intended for applications that want deterministic singleton settings at
// startup or tests that need a controlled clock. Existing standalone or
// worker/datacenter singleton generators are not modified.
func ConfigureDefaultSnowflake(opts ...SnowflakeOption) *Snowflake {
	cfg := applyDefaultSnowflakeOptions(opts)
	sf := newSnowflakeWithConfig(cfg)
	defaultSnowflake.Store(sf)
	return sf
}

// GetSnowflakeWithWorker returns a singleton Snowflake generator for workerID.
func GetSnowflakeWithWorker(workerID int64) *Snowflake {
	return GetSnowflakeWithWorkerWithOptions(workerID)
}

// GetSnowflakeWithWorkerWithOptions returns a singleton Snowflake generator for workerID using custom defaults.
func GetSnowflakeWithWorkerWithOptions(workerID int64, opts ...SnowflakeOption) *Snowflake {
	allOpts := append([]SnowflakeOption{WithSnowflakeWorkerID(workerID)}, opts...)
	cfg := applySnowflakeOptions(allOpts)
	if !cfg.datacenterSet {
		cfg.datacenterID = getDataCenterID(snowflakeMaxDatacenterID, cfg.interfaces)
	}
	return getCachedSnowflakeWithConfig(cfg)
}

// GetSnowflakeWithWorkerDataCenter returns a singleton Snowflake generator for worker/datacenter pair.
func GetSnowflakeWithWorkerDataCenter(workerID, datacenterID int64) *Snowflake {
	return GetSnowflakeWithWorkerDataCenterWithOptions(workerID, datacenterID)
}

// GetSnowflakeWithWorkerDataCenterWithOptions returns a singleton Snowflake generator for worker/datacenter pair using custom clock options.
func GetSnowflakeWithWorkerDataCenterWithOptions(workerID, datacenterID int64, opts ...SnowflakeOption) *Snowflake {
	allOpts := append([]SnowflakeOption{WithSnowflakeWorkerID(workerID), WithSnowflakeDatacenterID(datacenterID)}, opts...)
	return getCachedSnowflakeWithConfig(applySnowflakeOptions(allOpts))
}

func getCachedSnowflakeWithConfig(cfg snowflakeConfig) *Snowflake {
	cfg.workerID = normalizeSnowflakeID(cfg.workerID, snowflakeMaxWorkerID)
	cfg.datacenterID = normalizeSnowflakeID(cfg.datacenterID, snowflakeMaxDatacenterID)
	if !cfg.cache {
		return newSnowflakeWithConfig(cfg)
	}
	key := fmt.Sprintf("%d:%d", cfg.workerID, cfg.datacenterID)
	if v, ok := snowflakeCache.Load(key); ok {
		return v.(*Snowflake)
	}
	sf := newSnowflakeWithConfig(cfg)
	actual, _ := snowflakeCache.LoadOrStore(key, sf)
	return actual.(*Snowflake)
}

func newSnowflakeWithConfig(cfg snowflakeConfig) *Snowflake {
	workerID := normalizeSnowflakeID(cfg.workerID, snowflakeMaxWorkerID)
	datacenterID := normalizeSnowflakeID(cfg.datacenterID, snowflakeMaxDatacenterID)
	timeFunc := cfg.timeFunc
	if timeFunc == nil {
		timeFunc = currentMillis
	}
	tilNextMillis := cfg.tilNextMillis
	if tilNextMillis == nil {
		tilNextMillis = waitNextMillis
	}
	return &Snowflake{
		workerID:      workerID,
		datacenterID:  datacenterID,
		lastTimestamp: -1,
		timeFunc:      timeFunc,
		tilNextMillis: tilNextMillis,
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
	return GetDataCenterIDWithOptions(maxDatacenterID)
}

// GetDataCenterIDWithOptions derives a datacenter id using custom Snowflake providers.
func GetDataCenterIDWithOptions(maxDatacenterID int64, opts ...SnowflakeOption) int64 {
	return getDataCenterID(maxDatacenterID, applySnowflakeOptions(opts).interfaces)
}

func getDataCenterID(maxDatacenterID int64, interfaces func() ([]net.Interface, error)) int64 {
	if maxDatacenterID <= 0 {
		return 1
	}
	if maxDatacenterID == int64(^uint64(0)>>1) {
		maxDatacenterID--
	}
	for _, iface := range networkInterfaces(interfaces) {
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
	return GetWorkerIDWithOptions(datacenterID, maxWorkerID)
}

// GetWorkerIDWithOptions derives a worker id using custom Snowflake providers.
func GetWorkerIDWithOptions(datacenterID, maxWorkerID int64, opts ...SnowflakeOption) int64 {
	return getWorkerID(datacenterID, maxWorkerID, applySnowflakeOptions(opts).pid)
}

func getWorkerID(datacenterID, maxWorkerID int64, pid func() int) int64 {
	if maxWorkerID <= 0 {
		return 0
	}
	if pid == nil {
		pid = os.Getpid
	}
	h := fnv.New32a()
	_, _ = h.Write([]byte(strconv.FormatInt(datacenterID, 10)))
	_, _ = h.Write([]byte(strconv.Itoa(pid())))
	return int64(h.Sum32()&0xffff) % (maxWorkerID + 1)
}

// NanoId returns a default 21-character NanoId using a URL-safe alphabet.
func NanoId() string { return NanoIdN(21) }

// NanoIdN returns a NanoId with the specified length.
func NanoIdN(n int) string {
	return NanoIdNWithOptions(n)
}

// NanoIdNWithOptions returns a NanoId with the specified length and custom options.
func NanoIdNWithOptions(n int, opts ...NanoIDOption) string {
	return NanoIdWithOptions(append([]NanoIDOption{WithNanoIDLength(n)}, opts...)...)
}

// NanoIdWithOptions returns a NanoId using custom generation options.
func NanoIdWithOptions(opts ...NanoIDOption) string {
	cfg := applyNanoIDOptions(opts)
	n := cfg.length
	if n <= 0 {
		return ""
	}
	if len(cfg.alphabet) == 1 {
		return strings.Repeat(cfg.alphabet, n)
	}
	mask := byte(nextPowerOfTwo(len(cfg.alphabet)) - 1)
	step := (n*8 + 7) / 8
	out := make([]byte, 0, n)
	buf := make([]byte, step)
	for {
		fillRandomBytesWithConfig(cfg.randomConfig, buf)
		for i := 0; i < step && len(out) < n; i++ {
			idx := int(buf[i] & mask)
			if idx < len(cfg.alphabet) {
				out = append(out, cfg.alphabet[idx])
			}
		}
		if len(out) >= n {
			break
		}
	}
	return string(out[:n])
}

// GetSnowflakeNextID returns the next ID from the default singleton Snowflake generator.
func GetSnowflakeNextID() int64 { return GetSnowflakeNextIDWithOptions() }

// GetSnowflakeNextIDWithOptions returns the next ID from the default singleton Snowflake generator.
func GetSnowflakeNextIDWithOptions(opts ...SnowflakeOption) int64 {
	return GetSnowflakeWithOptions(opts...).NextID()
}

// GetSnowflakeNextIDStr returns the next ID string from the default singleton Snowflake generator.
func GetSnowflakeNextIDStr() string { return GetSnowflakeNextIDStrWithOptions() }

// GetSnowflakeNextIDStrWithOptions returns the next ID string from the default singleton Snowflake generator.
func GetSnowflakeNextIDStrWithOptions(opts ...SnowflakeOption) string {
	return GetSnowflakeWithOptions(opts...).NextIDStr()
}

func fillRandomBytesWithConfig(cfg randomConfig, buf []byte) {
	if cfg.reader == nil {
		cfg.reader = cryptorand.Reader
	}
	if _, err := io.ReadFull(cfg.reader, buf); err != nil {
		if cfg.fallbackSource != nil {
			for i := range buf {
				buf[i] = byte(cfg.fallbackSource.Intn(256)) // #nosec G115 -- Intn(256) always fits in byte.
			}
			return
		}
		// Compatibility fallback: ID helpers historically continued when the
		// primary entropy reader failed. This uses pseudo-random bytes and is not a
		// security boundary.
		defaultRandMu.Lock()
		defer defaultRandMu.Unlock()
		for i := range buf {
			buf[i] = byte(defaultRandLocked().Intn(256)) // #nosec G115 -- Intn(256) always fits in byte.
		}
	}
}

func defaultRandLocked() *mathrand.Rand {
	if defaultRand == nil {
		defaultRand = defaultRandProvider()
		if defaultRand == nil {
			defaultRand = newDefaultFallbackRand()
		}
	}
	return defaultRand
}

func newDefaultFallbackRand() *mathrand.Rand {
	return mathrand.New(mathrand.NewSource(time.Now().UnixNano())) // #nosec G404 -- fallback only for IDs when crypto/rand is unavailable.
}

func nextPowerOfTwo(n int) int {
	if n <= 1 {
		return 1
	}
	p := 1
	for p < n {
		p <<= 1
	}
	return p
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

func networkInterfaces(interfaces func() ([]net.Interface, error)) []net.Interface {
	if interfaces == nil {
		interfaces = net.Interfaces
	}
	ifaces, err := interfaces()
	if err != nil {
		return nil
	}
	return ifaces
}
