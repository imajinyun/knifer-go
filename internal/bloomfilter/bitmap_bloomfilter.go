package bloomfilter

// BitMapBloomFilter is a Bloom filter composed from multiple filters.
// It aggregates several BloomFilter instances and uses five different hash filters by default.
type BitMapBloomFilter struct {
	filters []BloomFilter
}

type bitMapBloomFilterConfig struct {
	m       int
	filters []BloomFilter
}

// BitMapBloomFilterOption customizes BitMapBloomFilter construction.
type BitMapBloomFilterOption func(*bitMapBloomFilterConfig)

// WithBitMapSize sets the M value in MB used by BitMapBloomFilter.
func WithBitMapSize(m int) BitMapBloomFilterOption {
	return func(c *bitMapBloomFilterConfig) { c.m = m }
}

// WithBloomFilters sets the Bloom filters aggregated by BitMapBloomFilter.
func WithBloomFilters(filters ...BloomFilter) BitMapBloomFilterOption {
	return func(c *bitMapBloomFilterConfig) {
		if len(filters) > 0 {
			c.filters = filters
		}
	}
}

// NewBitMapBloomFilterWithOptions creates a BitMapBloomFilter from functional options.
// WithBitMapSize is required. If WithBloomFilters is omitted, the default filter set is used.
func NewBitMapBloomFilterWithOptions(opts ...BitMapBloomFilterOption) *BitMapBloomFilter {
	cfg := bitMapBloomFilterConfig{}
	for _, opt := range opts {
		if opt != nil {
			opt(&cfg)
		}
	}
	return NewBitMapBloomFilterWithFilters(cfg.m, cfg.filters...)
}

// NewBitMapBloomFilter uses five default filters: Default, ELF, JS, PJW, and SDBM.
//
// m is the M value in MB and controls the underlying BitMap size. Final bits = m/5 * 1024 * 1024 * 8.
func NewBitMapBloomFilter(m int) *BitMapBloomFilter {
	mNum := int64(m) / 5
	size := mNum * 1024 * 1024 * 8
	return &BitMapBloomFilter{
		filters: []BloomFilter{
			NewDefaultFilter(size),
			NewELFFilter(size),
			NewJSFilter(size),
			NewPJWFilter(size),
			NewSDBMFilter(size),
		},
	}
}

// NewBitMapBloomFilterWithFilters creates a BitMapBloomFilter with custom filters.
// It keeps the utility toolkit-compatible m validation while replacing the default filter set.
func NewBitMapBloomFilterWithFilters(m int, filters ...BloomFilter) *BitMapBloomFilter {
	b := NewBitMapBloomFilter(m)
	if len(filters) > 0 {
		b.filters = filters
	}
	return b
}

// Add implements BloomFilter.Add. The value is considered added if any filter changes.
func (b *BitMapBloomFilter) Add(str string) bool {
	flag := false
	for _, f := range b.filters {
		if f.Add(str) {
			flag = true
		}
	}
	return flag
}

// Contains implements BloomFilter.Contains. All filters must report containment.
func (b *BitMapBloomFilter) Contains(str string) bool {
	for _, f := range b.filters {
		if !f.Contains(str) {
			return false
		}
	}
	return true
}
