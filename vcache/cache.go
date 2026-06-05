package vcache

import (
	"time"

	"github.com/imajinyun/go-knifer/internal/cache"
)

// Cache is a generic cache interface.
type Cache[K comparable, V any] interface {
	cache.Cache[K, V]
}

// CacheListener receives cache removal notifications.
type CacheListener[K comparable, V any] interface {
	cache.CacheListener[K, V]
}

// CacheListenerFunc adapts a function into CacheListener.
type CacheListenerFunc[K comparable, V any] func(key K, value V)

// OnRemove implements CacheListener.
func (f CacheListenerFunc[K, V]) OnRemove(key K, value V) { f(key, value) }

// CacheObj is a stored cache object.
type CacheObj[K comparable, V any] struct {
	*cache.CacheObj[K, V]
}

// Option customizes cache construction.
type Option[K comparable, V any] = cache.Option[K, V]

// FIFOCache is a first-in-first-out cache.
type FIFOCache[K comparable, V any] struct {
	*cache.FIFOCache[K, V]
}

// LFUCache is a least-frequently-used cache.
type LFUCache[K comparable, V any] struct {
	*cache.LFUCache[K, V]
}

// LRUCache is a least-recently-used cache.
type LRUCache[K comparable, V any] struct {
	*cache.LRUCache[K, V]
}

// NoCache is a cache implementation that stores nothing.
type NoCache[K comparable, V any] struct {
	*cache.NoCache[K, V]
}

// TimedCache is a cache with TTL support.
type TimedCache[K comparable, V any] struct {
	*cache.TimedCache[K, V]
}

// WeakCache is a weak-style timed cache.
type WeakCache[K comparable, V any] struct {
	*cache.WeakCache[K, V]
}

// Supplier supplies values lazily.
type Supplier[V any] func() (V, error)

// WithCapacity sets the maximum number of entries; 0 means unlimited.
func WithCapacity[K comparable, V any](capacity int) Option[K, V] {
	return cache.WithCapacity[K, V](capacity)
}

// WithTimeout sets the default entry expiration duration; 0 means no expiration.
func WithTimeout[K comparable, V any](timeout time.Duration) Option[K, V] {
	return cache.WithTimeout[K, V](timeout)
}

// WithListener sets the removal listener during cache construction.
func WithListener[K comparable, V any](listener CacheListener[K, V]) Option[K, V] {
	return cache.WithListener[K, V](listener)
}

// WithClock sets the time source used for cache expiration checks.
func WithClock[K comparable, V any](clock func() time.Time) Option[K, V] {
	return cache.WithClock[K, V](clock)
}

// NewFIFO creates a FIFO cache.
func NewFIFO[K comparable, V any](capacity int) *FIFOCache[K, V] {
	return &FIFOCache[K, V]{FIFOCache: cache.NewFIFO[K, V](capacity)}
}

// NewFIFOWithOptions creates a FIFO cache customized by options.
func NewFIFOWithOptions[K comparable, V any](opts ...Option[K, V]) *FIFOCache[K, V] {
	return &FIFOCache[K, V]{FIFOCache: cache.NewFIFOWithOptions[K, V](opts...)}
}

// NewFIFOWithTimeout creates a FIFO cache with timeout.
func NewFIFOWithTimeout[K comparable, V any](capacity int, timeout time.Duration) *FIFOCache[K, V] {
	return &FIFOCache[K, V]{FIFOCache: cache.NewFIFOWithTimeout[K, V](capacity, timeout)}
}

// NewLFU creates an LFU cache.
func NewLFU[K comparable, V any](capacity int) *LFUCache[K, V] {
	return &LFUCache[K, V]{LFUCache: cache.NewLFU[K, V](capacity)}
}

// NewLFUWithOptions creates an LFU cache customized by options.
func NewLFUWithOptions[K comparable, V any](opts ...Option[K, V]) *LFUCache[K, V] {
	return &LFUCache[K, V]{LFUCache: cache.NewLFUWithOptions[K, V](opts...)}
}

// NewLFUWithTimeout creates an LFU cache with timeout.
func NewLFUWithTimeout[K comparable, V any](capacity int, timeout time.Duration) *LFUCache[K, V] {
	return &LFUCache[K, V]{LFUCache: cache.NewLFUWithTimeout[K, V](capacity, timeout)}
}

// NewLRU creates an LRU cache.
func NewLRU[K comparable, V any](capacity int) *LRUCache[K, V] {
	return &LRUCache[K, V]{LRUCache: cache.NewLRU[K, V](capacity)}
}

// NewLRUWithOptions creates an LRU cache customized by options.
func NewLRUWithOptions[K comparable, V any](opts ...Option[K, V]) *LRUCache[K, V] {
	return &LRUCache[K, V]{LRUCache: cache.NewLRUWithOptions[K, V](opts...)}
}

// NewLRUWithTimeout creates an LRU cache with timeout.
func NewLRUWithTimeout[K comparable, V any](capacity int, timeout time.Duration) *LRUCache[K, V] {
	return &LRUCache[K, V]{LRUCache: cache.NewLRUWithTimeout[K, V](capacity, timeout)}
}

// NewNoCache creates a no-op cache.
func NewNoCache[K comparable, V any]() *NoCache[K, V] {
	return &NoCache[K, V]{NoCache: cache.NewNoCache[K, V]()}
}

// NewTimed creates a timed cache.
func NewTimed[K comparable, V any](timeout time.Duration) *TimedCache[K, V] {
	return &TimedCache[K, V]{TimedCache: cache.NewTimed[K, V](timeout)}
}

// NewTimedWithOptions creates a timed cache customized by options.
func NewTimedWithOptions[K comparable, V any](opts ...Option[K, V]) *TimedCache[K, V] {
	return &TimedCache[K, V]{TimedCache: cache.NewTimedWithOptions[K, V](opts...)}
}

// NewTimedScheduled creates a timed cache with scheduled pruning.
func NewTimedScheduled[K comparable, V any](timeout, schedulePruneDelay time.Duration) *TimedCache[K, V] {
	return &TimedCache[K, V]{TimedCache: cache.NewTimedScheduled[K, V](timeout, schedulePruneDelay)}
}

// NewWeak creates a weak-style timed cache.
func NewWeak[K comparable, V any](timeout time.Duration) *WeakCache[K, V] {
	return &WeakCache[K, V]{WeakCache: cache.NewWeak[K, V](timeout)}
}

// NewWeakWithOptions creates a weak-style timed cache customized by options.
func NewWeakWithOptions[K comparable, V any](opts ...Option[K, *V]) *WeakCache[K, V] {
	return &WeakCache[K, V]{WeakCache: cache.NewWeakWithOptions[K, V](opts...)}
}
