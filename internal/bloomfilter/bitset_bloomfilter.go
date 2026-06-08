package bloomfilter

import (
	"bufio"
	"io"
	"math"
	"os"
)

// BitSetBloomFilter is a fixed-size bitset based Bloom filter.
// Hash algorithms are used in a fixed order; only the algorithm count is configurable.
type BitSetBloomFilter struct {
	bits               []uint64 // Simulates BitSet.
	bitSetSize         int
	addedElements      int
	hashFunctionNumber int
}

type bitSetBloomFilterConfig struct {
	c int
	n int
	k int
}

type bloomFileConfig struct {
	openFile func(string) (io.ReadCloser, error)
}

// BitSetBloomFilterOption customizes BitSetBloomFilter construction.
type BitSetBloomFilterOption func(*bitSetBloomFilterConfig)

// FileOption customizes bloom filter file helpers.
type FileOption func(*bloomFileConfig)

// WithOpenFile sets the file opener used by bloom filter file helpers.
func WithOpenFile(openFile func(string) (io.ReadCloser, error)) FileOption {
	return func(c *bloomFileConfig) { c.openFile = openFile }
}

func applyFileOptions(opts []FileOption) bloomFileConfig {
	cfg := bloomFileConfig{openFile: defaultOpenFile}
	for _, opt := range opts {
		if opt != nil {
			opt(&cfg)
		}
	}
	if cfg.openFile == nil {
		cfg.openFile = defaultOpenFile
	}
	return cfg
}

func defaultOpenFile(path string) (io.ReadCloser, error) { return os.Open(path) }

// WithBitSetCapacity sets the preallocated maximum record count.
func WithBitSetCapacity(c int) BitSetBloomFilterOption {
	return func(cfg *bitSetBloomFilterConfig) { cfg.c = c }
}

// WithExpectedElements sets the expected record count.
func WithExpectedElements(n int) BitSetBloomFilterOption {
	return func(cfg *bitSetBloomFilterConfig) { cfg.n = n }
}

// WithHashFunctionNumber sets the number of hash functions, in range [1, 8].
func WithHashFunctionNumber(k int) BitSetBloomFilterOption {
	return func(cfg *bitSetBloomFilterConfig) { cfg.k = k }
}

// NewBitSetBloomFilterWithOptions creates a BitSetBloomFilter from functional options.
// WithBitSetCapacity, WithExpectedElements, and WithHashFunctionNumber are required.
func NewBitSetBloomFilterWithOptions(opts ...BitSetBloomFilterOption) *BitSetBloomFilter {
	bf, _ := NewBitSetBloomFilterWithOptionsE(opts...)
	return bf
}

// NewBitSetBloomFilterWithOptionsE creates a BitSetBloomFilter from functional options and returns validation errors.
func NewBitSetBloomFilterWithOptionsE(opts ...BitSetBloomFilterOption) (*BitSetBloomFilter, error) {
	cfg := bitSetBloomFilterConfig{}
	for _, opt := range opts {
		if opt != nil {
			opt(&cfg)
		}
	}
	return newBitSetBloomFilterWithConfig(cfg)
}

// NewBitSetBloomFilter creates a Bloom filter with c*k bits.
//
// c is the preallocated maximum record count, typically twice the expected inserted count.
// n is the expected record count.
// k is the number of hash functions, in range [1, 8].
func NewBitSetBloomFilter(c, n, k int) *BitSetBloomFilter {
	bf, _ := NewBitSetBloomFilterE(c, n, k)
	return bf
}

// NewBitSetBloomFilterE creates a Bloom filter with c*k bits and returns validation errors.
func NewBitSetBloomFilterE(c, n, k int) (*BitSetBloomFilter, error) {
	return NewBitSetBloomFilterWithOptionsE(WithBitSetCapacity(c), WithExpectedElements(n), WithHashFunctionNumber(k))
}

func newBitSetBloomFilterWithConfig(cfg bitSetBloomFilterConfig) (*BitSetBloomFilter, error) {
	if cfg.c <= 0 {
		return nil, invalidInputf("parameter c must be positive")
	}
	if cfg.n <= 0 {
		return nil, invalidInputf("parameter n must be positive")
	}
	if cfg.k < 1 || cfg.k > 8 {
		return nil, invalidInputf("hash function number must be between 1 and 8")
	}
	size := cfg.c * cfg.k
	return &BitSetBloomFilter{
		bits:               make([]uint64, (size+63)/64),
		bitSetSize:         size,
		addedElements:      cfg.n,
		hashFunctionNumber: cfg.k,
	}, nil
}

func (b *BitSetBloomFilter) setBit(pos int) { b.bits[pos>>6] |= 1 << uint(pos&63) }

func (b *BitSetBloomFilter) getBit(pos int) bool {
	return (b.bits[pos>>6]>>uint(pos&63))&1 == 1
}

// InitFromFile initializes the filter from a file by adding each line.
func (b *BitSetBloomFilter) InitFromFile(path string) error {
	return b.InitFromFileWithOptions(path)
}

// InitFromFileWithOptions initializes the filter from a file using options.
func (b *BitSetBloomFilter) InitFromFileWithOptions(path string, opts ...FileOption) error {
	f, err := applyFileOptions(opts).openFile(path)
	if err != nil {
		return wrapBloomFilterIO("open bloom filter file "+path, err)
	}
	defer func() { _ = f.Close() }()
	return b.InitFromReader(f)
}

// InitFromReader initializes the filter from a reader by adding each line.
func (b *BitSetBloomFilter) InitFromReader(reader io.Reader) error {
	r := bufio.NewReader(reader)
	for {
		line, err := r.ReadString('\n')
		if len(line) > 0 {
			// Trim trailing line endings.
			for len(line) > 0 && (line[len(line)-1] == '\n' || line[len(line)-1] == '\r') {
				line = line[:len(line)-1]
			}
			b.Add(line)
		}
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return wrapBloomFilterIO("read bloom filter data", err)
		}
	}
}

// Add inserts a string and returns false if it likely already exists.
func (b *BitSetBloomFilter) Add(str string) bool {
	if b.Contains(str) {
		return false
	}
	positions := b.createHashes(str, b.hashFunctionNumber)
	for _, v := range positions {
		pos := absInt(v % int32(b.bitSetSize))
		b.setBit(int(pos))
	}
	return true
}

// Contains reports whether the string may exist.
func (b *BitSetBloomFilter) Contains(str string) bool {
	positions := b.createHashes(str, b.hashFunctionNumber)
	for _, v := range positions {
		pos := absInt(v % int32(b.bitSetSize))
		if !b.getBit(int(pos)) {
			return false
		}
	}
	return true
}

// FalsePositiveProbability returns the current false positive probability: (1 - e^(-k * n / m)) ^ k.
func (b *BitSetBloomFilter) FalsePositiveProbability() float64 {
	return math.Pow(1-math.Exp(-float64(b.hashFunctionNumber)*float64(b.addedElements)/float64(b.bitSetSize)),
		float64(b.hashFunctionNumber))
}

// createHashes returns multiple hash values.
func (b *BitSetBloomFilter) createHashes(str string, hashNumber int) []int32 {
	out := make([]int32, hashNumber)
	for i := 0; i < hashNumber; i++ {
		out[i] = bitSetHash(str, i)
	}
	return out
}

// bitSetHash matches the utility toolkit BitSetBloomFilter.hash.
func bitSetHash(str string, k int) int32 {
	switch k {
	case 0:
		return RsHash(str)
	case 1:
		return JsHash(str)
	case 2:
		return ElfHash(str)
	case 3:
		return BkdrHash(str)
	case 4:
		return ApHash(str)
	case 5:
		return DjbHash(str)
	case 6:
		return SdbmHash(str)
	case 7:
		return PjwHash(str)
	default:
		return 0
	}
}

func absInt(v int32) int32 {
	if v < 0 {
		return -v
	}
	return v
}
