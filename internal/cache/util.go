package cache

import "time"

// Package-level constructors similar to hutool-cache CacheUtil.

// NewFIFO creates a FIFO cache.
func NewFIFO[K comparable, V any](capacity int) *FIFOCache[K, V] {
	return NewFIFOCache[K, V](capacity)
}

// NewFIFOWithTimeout creates a FIFO cache with a default timeout.
func NewFIFOWithTimeout[K comparable, V any](capacity int, timeout time.Duration) *FIFOCache[K, V] {
	return NewFIFOCacheWithTimeout[K, V](capacity, timeout)
}

// NewLFU creates an LFU cache.
func NewLFU[K comparable, V any](capacity int) *LFUCache[K, V] {
	return NewLFUCache[K, V](capacity)
}

// NewLFUWithTimeout creates an LFU cache with a default timeout.
func NewLFUWithTimeout[K comparable, V any](capacity int, timeout time.Duration) *LFUCache[K, V] {
	return NewLFUCacheWithTimeout[K, V](capacity, timeout)
}

// NewLRU creates an LRU cache.
func NewLRU[K comparable, V any](capacity int) *LRUCache[K, V] {
	return NewLRUCache[K, V](capacity)
}

// NewLRUWithTimeout creates an LRU cache with a default timeout.
func NewLRUWithTimeout[K comparable, V any](capacity int, timeout time.Duration) *LRUCache[K, V] {
	return NewLRUCacheWithTimeout[K, V](capacity, timeout)
}

// NewTimed creates a timed cache.
func NewTimed[K comparable, V any](timeout time.Duration) *TimedCache[K, V] {
	return NewTimedCache[K, V](timeout)
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

// NewNo creates a no-op cache.
func NewNo[K comparable, V any]() *NoCache[K, V] {
	return NewNoCache[K, V]()
}
