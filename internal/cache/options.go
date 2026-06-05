package cache

import "time"

type cacheConfig[K comparable, V any] struct {
	capacity int
	timeout  time.Duration
	listener CacheListener[K, V]
	clock    func() time.Time
}

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
	return func(c *cacheConfig[K, V]) { c.listener = listener }
}

// WithClock sets the time source used for cache expiration checks.
func WithClock[K comparable, V any](clock func() time.Time) Option[K, V] {
	return func(c *cacheConfig[K, V]) { c.clock = clock }
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
