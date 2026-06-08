package bloomfilter

// CreateBitSet creates a BitSet-based Bloom filter.
// See NewBitSetBloomFilter for detailed parameter semantics.
func CreateBitSet(c, n, k int) *BitSetBloomFilter { return NewBitSetBloomFilter(c, n, k) }

// CreateBitSetE creates a BitSet-based Bloom filter and returns validation errors.
func CreateBitSetE(c, n, k int) (*BitSetBloomFilter, error) { return NewBitSetBloomFilterE(c, n, k) }

// CreateBitMap creates a BitMap-backed Bloom filter.
func CreateBitMap(m int) *BitMapBloomFilter { return NewBitMapBloomFilter(m) }

// CreateBitMapE creates a BitMap-backed Bloom filter and returns validation errors.
func CreateBitMapE(m int) (*BitMapBloomFilter, error) { return NewBitMapBloomFilterE(m) }
