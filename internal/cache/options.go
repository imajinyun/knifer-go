package cache

import "time"

type cacheConfig[K comparable, V any] struct {
	capacity      int
	timeout       time.Duration
	listener      CacheListener[K, V]
	clock         func() time.Time
	tickerFactory TickerFactory
	runner        func(func())
	finalizer     any
	finalizerOff  bool
}

// Ticker stops a scheduled cache pruning ticker created by TickerFactory.
type Ticker interface {
	Stop()
}

// TickerFactory creates a ticker channel and stopper for scheduled pruning.
type TickerFactory func(time.Duration) (<-chan time.Time, Ticker)

// Option customizes cache construction.
type Option[K comparable, V any] func(*cacheConfig[K, V])

// WithCapacity sets the maximum number of entries; 0 means unlimited.
func WithCapacity[K comparable, V any](capacity int) Option[K, V] {
	return func(c *cacheConfig[K, V]) { c.capacity = capacity }
}

// WithTimeout sets the default entry expiration duration; 0 means no expiration.
func WithTimeout[K comparable, V any](timeout time.Duration) Option[K, V] {
	return func(c *cacheConfig[K, V]) { c.timeout = timeout }
}

// WithListener sets the removal listener during cache construction.
func WithListener[K comparable, V any](listener CacheListener[K, V]) Option[K, V] {
	return func(c *cacheConfig[K, V]) {
		if listener != nil {
			c.listener = listener
		}
	}
}

// WithClock sets the time source used for cache expiration checks.
func WithClock[K comparable, V any](clock func() time.Time) Option[K, V] {
	return func(c *cacheConfig[K, V]) {
		if clock != nil {
			c.clock = clock
		}
	}
}

// WithTickerFactory sets the ticker factory used by scheduled pruning.
func WithTickerFactory[K comparable, V any](factory TickerFactory) Option[K, V] {
	return func(c *cacheConfig[K, V]) {
		if factory != nil {
			c.tickerFactory = factory
		}
	}
}

// WithRunner sets the runner used by scheduled pruning tasks.
func WithRunner[K comparable, V any](runner func(func())) Option[K, V] {
	return func(c *cacheConfig[K, V]) {
		if runner != nil {
			c.runner = runner
		}
	}
}

// WithWeakFinalizerFunc sets the finalizer provider used by WeakCache.
func WithWeakFinalizerFunc[K comparable, V any](finalizer func(*V, func(*V))) Option[K, *V] {
	return func(c *cacheConfig[K, *V]) {
		if finalizer != nil {
			c.finalizer = finalizer
		}
	}
}

// WithWeakFinalizerEnabled controls whether WeakCache registers GC finalizers.
func WithWeakFinalizerEnabled[K comparable, V any](enabled bool) Option[K, *V] {
	return func(c *cacheConfig[K, *V]) { c.finalizerOff = !enabled }
}

func applyOptions[K comparable, V any](opts []Option[K, V]) cacheConfig[K, V] {
	cfg := cacheConfig[K, V]{}
	for _, opt := range opts {
		if opt != nil {
			opt(&cfg)
		}
	}
	return cfg
}

func applyListener[K comparable, V any](c *abstractCache[K, V], listener CacheListener[K, V]) {
	if listener != nil {
		c.listener = listener
	}
}

func applyClock[K comparable, V any](c *abstractCache[K, V], clock func() time.Time) {
	c.setClock(clock)
}

func applyTickerFactory[K comparable, V any](c *abstractCache[K, V], factory TickerFactory) {
	c.setTickerFactory(factory)
}

func applyRunner[K comparable, V any](c *abstractCache[K, V], runner func(func())) {
	c.setRunner(runner)
}

func newTicker(delay time.Duration) (<-chan time.Time, Ticker) {
	ticker := time.NewTicker(delay)
	return ticker.C, ticker
}

func defaultRunner(fn func()) { go fn() }
