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

const (
	// BloomMachine32 uses 32-bit bitmap words.
	BloomMachine32 = bloomfilter.Machine32
	// BloomMachine64 uses 64-bit bitmap words.
	BloomMachine64 = bloomfilter.Machine64
)

// NewBitMapBloomFilter creates a bitmap bloom filter.
func NewBitMapBloomFilter(m int) *BitMapBloomFilter { return bloomfilter.NewBitMapBloomFilter(m) }

// NewBitMapBloomFilterWithFilters creates a bitmap bloom filter with filters.
func NewBitMapBloomFilterWithFilters(m int, filters ...BloomFilter) *BitMapBloomFilter {
	return bloomfilter.NewBitMapBloomFilterWithFilters(m, filters...)
}

// NewBitSetBloomFilter creates a bitset bloom filter.
func NewBitSetBloomFilter(c, n, k int) *BitSetBloomFilter {
	return bloomfilter.NewBitSetBloomFilter(c, n, k)
}

// NewFuncFilter creates a function-backed bloom filter.
func NewFuncFilter(maxValue int64, hashFunc HashFunc) *FuncFilter {
	return bloomfilter.NewFuncFilter(maxValue, hashFunc)
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
