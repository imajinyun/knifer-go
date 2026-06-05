package vblf

import "github.com/imajinyun/go-knifer/internal/bloomfilter"

// BitMap is the bitmap abstraction used by bloom filters.
type BitMap = bloomfilter.BitMap

// BloomFilter is the bloom filter interface.
type BloomFilter = bloomfilter.BloomFilter

// HashFunc calculates a hash value for a string.
type HashFunc = bloomfilter.HashFunc

// FuncFilter is a hash-function-backed bloom filter.
type FuncFilter = bloomfilter.FuncFilter

// BitMapBloomFilter combines multiple filters over a bitmap.
type BitMapBloomFilter = bloomfilter.BitMapBloomFilter

// BitSetBloomFilter is a bitset-backed bloom filter.
type BitSetBloomFilter = bloomfilter.BitSetBloomFilter

// IntMap is an int bitmap implementation.
type IntMap = bloomfilter.IntMap

// LongMap is a long bitmap implementation.
type LongMap = bloomfilter.LongMap

// Error is the code-aware error type returned by bloom filter helpers.
type Error = bloomfilter.BloomFilterError

// FuncFilterOption customizes FuncFilter construction.
type FuncFilterOption = bloomfilter.FuncFilterOption

// BitMapBloomFilterOption customizes BitMapBloomFilter construction.
type BitMapBloomFilterOption = bloomfilter.BitMapBloomFilterOption

// BitSetBloomFilterOption customizes BitSetBloomFilter construction.
type BitSetBloomFilterOption = bloomfilter.BitSetBloomFilterOption

const (
	// BloomMachine32 uses 32-bit bitmap words.
	BloomMachine32 = bloomfilter.Machine32
	// BloomMachine64 uses 64-bit bitmap words.
	BloomMachine64 = bloomfilter.Machine64
)

// NewBitMapBloomFilter creates a bitmap bloom filter.
func NewBitMapBloomFilter(m int) *BitMapBloomFilter { return bloomfilter.NewBitMapBloomFilter(m) }

// NewBitMapBloomFilterWithOptions creates a bitmap bloom filter with options.
func NewBitMapBloomFilterWithOptions(opts ...BitMapBloomFilterOption) *BitMapBloomFilter {
	return bloomfilter.NewBitMapBloomFilterWithOptions(opts...)
}

// NewBitMapBloomFilterWithFilters creates a bitmap bloom filter with filters.
func NewBitMapBloomFilterWithFilters(m int, filters ...BloomFilter) *BitMapBloomFilter {
	return bloomfilter.NewBitMapBloomFilterWithFilters(m, filters...)
}

// NewBitSetBloomFilter creates a bitset bloom filter.
func NewBitSetBloomFilter(c, n, k int) *BitSetBloomFilter {
	return bloomfilter.NewBitSetBloomFilter(c, n, k)
}

// NewBitSetBloomFilterWithOptions creates a bitset bloom filter with options.
func NewBitSetBloomFilterWithOptions(opts ...BitSetBloomFilterOption) *BitSetBloomFilter {
	return bloomfilter.NewBitSetBloomFilterWithOptions(opts...)
}

// NewFuncFilter creates a function-backed bloom filter.
func NewFuncFilter(maxValue int64, hashFunc HashFunc) *FuncFilter {
	return bloomfilter.NewFuncFilter(maxValue, hashFunc)
}

// NewFuncFilterWithOptions creates a function-backed bloom filter with options.
func NewFuncFilterWithOptions(opts ...FuncFilterOption) *FuncFilter {
	return bloomfilter.NewFuncFilterWithOptions(opts...)
}

// WithMaxValue sets the maximum hash value range for FuncFilter.
func WithMaxValue(maxValue int64) FuncFilterOption { return bloomfilter.WithMaxValue(maxValue) }

// WithMachineNum sets the backing bitmap machine word size for FuncFilter.
func WithMachineNum(machineNum int) FuncFilterOption { return bloomfilter.WithMachineNum(machineNum) }

// WithHashFunc sets the hash function used by FuncFilter.
func WithHashFunc(hashFunc HashFunc) FuncFilterOption { return bloomfilter.WithHashFunc(hashFunc) }

// WithBitMapSize sets the M value in MB used by BitMapBloomFilter.
func WithBitMapSize(m int) BitMapBloomFilterOption { return bloomfilter.WithBitMapSize(m) }

// WithBloomFilters sets the Bloom filters aggregated by BitMapBloomFilter.
func WithBloomFilters(filters ...BloomFilter) BitMapBloomFilterOption {
	return bloomfilter.WithBloomFilters(filters...)
}

// WithBitSetCapacity sets the preallocated maximum record count.
func WithBitSetCapacity(c int) BitSetBloomFilterOption { return bloomfilter.WithBitSetCapacity(c) }

// WithExpectedElements sets the expected record count.
func WithExpectedElements(n int) BitSetBloomFilterOption { return bloomfilter.WithExpectedElements(n) }

// WithHashFunctionNumber sets the number of hash functions, in range [1, 8].
func WithHashFunctionNumber(k int) BitSetBloomFilterOption {
	return bloomfilter.WithHashFunctionNumber(k)
}

// NewDefaultBloomFilter creates a default bloom filter.
func NewDefaultBloomFilter(maxValue int64) *FuncFilter { return bloomfilter.NewDefaultFilter(maxValue) }

// BloomRSHash returns RS hash.
func BloomRSHash(str string) int32 { return bloomfilter.RsHash(str) }

// BloomJSHash returns JS hash.
func BloomJSHash(str string) int32 { return bloomfilter.JsHash(str) }

// BloomELFHash returns ELF hash.
func BloomELFHash(str string) int32 { return bloomfilter.ElfHash(str) }

// BloomBKDRHash returns BKDR hash.
func BloomBKDRHash(str string) int32 { return bloomfilter.BkdrHash(str) }

// BloomSDBMHash returns SDBM hash.
func BloomSDBMHash(str string) int32 { return bloomfilter.SdbmHash(str) }

// BloomDJBHash returns DJB hash.
func BloomDJBHash(str string) int32 { return bloomfilter.DjbHash(str) }

// BloomFNVHash returns FNV hash.
func BloomFNVHash(str string) int32 { return bloomfilter.FnvHashString(str) }
