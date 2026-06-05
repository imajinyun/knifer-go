package cache

import "time"

// NewFIFO creates a FIFO cache.
func NewFIFO[K comparable, V any](capacity int) *FIFOCache[K, V] {
	return NewFIFOCache[K, V](capacity)
}

// NewFIFOWithOptions creates a FIFO cache customized by options.
func NewFIFOWithOptions[K comparable, V any](opts ...Option[K, V]) *FIFOCache[K, V] {
	return NewFIFOCacheWithOptions[K, V](opts...)
}

// NewFIFOWithTimeout creates a FIFO cache with a default timeout.
func NewFIFOWithTimeout[K comparable, V any](capacity int, timeout time.Duration) *FIFOCache[K, V] {
	return NewFIFOCacheWithTimeout[K, V](capacity, timeout)
}

// NewLFU creates an LFU cache.
func NewLFU[K comparable, V any](capacity int) *LFUCache[K, V] {
	return NewLFUCache[K, V](capacity)
}

// NewLFUWithOptions creates an LFU cache customized by options.
func NewLFUWithOptions[K comparable, V any](opts ...Option[K, V]) *LFUCache[K, V] {
	return NewLFUCacheWithOptions[K, V](opts...)
}

// NewLFUWithTimeout creates an LFU cache with a default timeout.
func NewLFUWithTimeout[K comparable, V any](capacity int, timeout time.Duration) *LFUCache[K, V] {
	return NewLFUCacheWithTimeout[K, V](capacity, timeout)
}

// NewLRU creates an LRU cache.
func NewLRU[K comparable, V any](capacity int) *LRUCache[K, V] {
	return NewLRUCache[K, V](capacity)
}

// NewLRUWithOptions creates an LRU cache customized by options.
func NewLRUWithOptions[K comparable, V any](opts ...Option[K, V]) *LRUCache[K, V] {
	return NewLRUCacheWithOptions[K, V](opts...)
}

// NewLRUWithTimeout creates an LRU cache with a default timeout.
func NewLRUWithTimeout[K comparable, V any](capacity int, timeout time.Duration) *LRUCache[K, V] {
	return NewLRUCacheWithTimeout[K, V](capacity, timeout)
}

// NewTimed creates a timed cache.
func NewTimed[K comparable, V any](timeout time.Duration) *TimedCache[K, V] {
	return NewTimedCache[K, V](timeout)
}

// NewTimedWithOptions creates a timed cache customized by options.
func NewTimedWithOptions[K comparable, V any](opts ...Option[K, V]) *TimedCache[K, V] {
	return NewTimedCacheWithOptions[K, V](opts...)
}

// NewTimedScheduled creates a timed cache and starts background pruning.
func NewTimedScheduled[K comparable, V any](timeout, schedulePruneDelay time.Duration) *TimedCache[K, V] {
	c := NewTimedCache[K, V](timeout)
	c.SchedulePrune(schedulePruneDelay)
	return c
}

// NewWeak creates a weak-reference-like cache.
func NewWeak[K comparable, V any](timeout time.Duration) *WeakCache[K, V] {
	return NewWeakCache[K, V](timeout)
}

// NewWeakWithOptions creates a weak-reference-like cache customized by options.
func NewWeakWithOptions[K comparable, V any](opts ...Option[K, *V]) *WeakCache[K, V] {
	return NewWeakCacheWithOptions[K, V](opts...)
}

// NewNo creates a no-op cache.
func NewNo[K comparable, V any]() *NoCache[K, V] {
	return NewNoCache[K, V]()
}
