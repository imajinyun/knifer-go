package bloomfilter

import "fmt"

// BloomFilter is the Bloom filter interface.
type BloomFilter interface {
	// Contains reports whether the string may exist in the filter.
	Contains(str string) bool
	// Add inserts a string into the filter. It returns false if the value may already exist.
	Add(str string) bool
}

// HashFunc calculates a hash value for FuncFilter.
type HashFunc func(str string) int64

// FuncFilter is a Bloom filter backed by a custom hash function.
type FuncFilter struct {
	bm       BitMap
	size     int64
	hashFunc HashFunc
}

type funcFilterConfig struct {
	maxValue   int64
	machineNum int
	hashFunc   HashFunc
}

// FuncFilterOption customizes FuncFilter construction.
type FuncFilterOption func(*funcFilterConfig)

// WithMaxValue sets the maximum hash value range for FuncFilter.
func WithMaxValue(maxValue int64) FuncFilterOption {
	return func(c *funcFilterConfig) { c.maxValue = maxValue }
}

// WithMachineNum sets the backing bitmap machine word size for FuncFilter.
func WithMachineNum(machineNum int) FuncFilterOption {
	return func(c *funcFilterConfig) { c.machineNum = machineNum }
}

// WithHashFunc sets the hash function used by FuncFilter.
func WithHashFunc(hashFunc HashFunc) FuncFilterOption {
	return func(c *funcFilterConfig) {
		if hashFunc != nil {
			c.hashFunc = hashFunc
		}
	}
}

// DefaultMachineNum is the default machine word size for FuncFilter.
var DefaultMachineNum = Machine32

func defaultFuncFilterConfig() funcFilterConfig {
	return funcFilterConfig{
		machineNum: DefaultMachineNum,
		hashFunc:   func(s string) int64 { return int64(JavaDefaultHash(s)) },
	}
}

// NewFuncFilterWithOptions creates a FuncFilter from functional options.
// WithMaxValue is required. If WithHashFunc is omitted, JavaDefaultHash is used.
func NewFuncFilterWithOptions(opts ...FuncFilterOption) *FuncFilter {
	cfg := defaultFuncFilterConfig()
	for _, opt := range opts {
		if opt != nil {
			opt(&cfg)
		}
	}
	return NewFuncFilterWithMachineNum(cfg.maxValue, cfg.machineNum, cfg.hashFunc)
}

// NewFuncFilter creates a FuncFilter with the default machine word size.
func NewFuncFilter(maxValue int64, hashFunc HashFunc) *FuncFilter {
	return NewFuncFilterWithMachineNum(maxValue, DefaultMachineNum, hashFunc)
}

// NewFuncFilterWithMachineNum creates a FuncFilter with the specified machine word size.
func NewFuncFilterWithMachineNum(maxValue int64, machineNum int, hashFunc HashFunc) *FuncFilter {
	if maxValue < 1 || maxValue > 0x7FFFFFFF {
		panic(fmt.Sprintf("maxValue must be between 1 and %d", int64(0x7FFFFFFF)))
	}
	capacity := int((maxValue + int64(machineNum) - 1) / int64(machineNum))
	var bm BitMap
	switch machineNum {
	case Machine32:
		bm = NewIntMap(capacity)
	case Machine64:
		bm = NewLongMap(capacity)
	default:
		panic("Error Machine number!")
	}
	return &FuncFilter{bm: bm, size: maxValue, hashFunc: hashFunc}
}

// hash calls the underlying hash function, applies modulo size, and returns an absolute value.
func (f *FuncFilter) hash(str string) int64 {
	v := f.hashFunc(str) % f.size
	if v < 0 {
		v = -v
	}
	return v
}

// Contains implements BloomFilter.Contains.
func (f *FuncFilter) Contains(str string) bool { return f.bm.Contains(f.hash(str)) }

// Add implements BloomFilter.Add.
func (f *FuncFilter) Add(str string) bool {
	h := f.hash(str)
	if f.bm.Contains(h) {
		return false
	}
	f.bm.Add(h)
	return true
}

// ============= Convenient filters based on specific hash algorithms =============

// NewDefaultFilter creates a default Bloom filter using Java String.hashCode.
func NewDefaultFilter(maxValue int64) *FuncFilter {
	return NewFuncFilter(maxValue, func(s string) int64 { return int64(JavaDefaultHash(s)) })
}

// NewELFFilter creates an ELF hash filter.
func NewELFFilter(maxValue int64) *FuncFilter {
	return NewFuncFilter(maxValue, func(s string) int64 { return int64(ElfHash(s)) })
}

// NewFNVFilter creates an FNV hash filter.
func NewFNVFilter(maxValue int64) *FuncFilter {
	return NewFuncFilter(maxValue, func(s string) int64 { return int64(FnvHashString(s)) })
}

// NewHfFilter creates an HF hash filter.
func NewHfFilter(maxValue int64) *FuncFilter {
	return NewFuncFilter(maxValue, HfHash)
}

// NewHfIpFilter creates an HFIP hash filter.
func NewHfIpFilter(maxValue int64) *FuncFilter {
	return NewFuncFilter(maxValue, HfIpHash)
}

// NewJSFilter creates a JS hash filter.
func NewJSFilter(maxValue int64) *FuncFilter {
	return NewFuncFilter(maxValue, func(s string) int64 { return int64(JsHash(s)) })
}

// NewPJWFilter creates a PJW hash filter.
func NewPJWFilter(maxValue int64) *FuncFilter {
	return NewFuncFilter(maxValue, func(s string) int64 { return int64(PjwHash(s)) })
}

// NewRSFilter creates an RS hash filter.
func NewRSFilter(maxValue int64) *FuncFilter {
	return NewFuncFilter(maxValue, func(s string) int64 { return int64(RsHash(s)) })
}

// NewSDBMFilter creates an SDBM hash filter.
func NewSDBMFilter(maxValue int64) *FuncFilter {
	return NewFuncFilter(maxValue, func(s string) int64 { return int64(SdbmHash(s)) })
}

// NewTianlFilter creates a TianL hash filter.
func NewTianlFilter(maxValue int64) *FuncFilter {
	return NewFuncFilter(maxValue, TianlHash)
}
