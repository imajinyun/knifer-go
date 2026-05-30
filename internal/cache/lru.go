package cache

import "time"

// LRUCache evicts the least recently used entry when capacity is exceeded.
type LRUCache[K comparable, V any] struct {
	abstractCache[K, V]
}

// NewLRUCache creates an LRU cache with no default timeout.
func NewLRUCache[K comparable, V any](capacity int) *LRUCache[K, V] {
	return NewLRUCacheWithTimeout[K, V](capacity, 0)
}

// NewLRUCacheWithTimeout creates an LRU cache with a default timeout.
func NewLRUCacheWithTimeout[K comparable, V any](capacity int, timeout time.Duration) *LRUCache[K, V] {
	c := &LRUCache[K, V]{}
	c.init(capacity, timeout, lruPrune[K, V])
	c.moveToBackOnGet = true
	return c
}

// SetListener sets the removal listener and returns the cache for chaining.
func (c *LRUCache[K, V]) SetListener(l CacheListener[K, V]) Cache[K, V] {
	c.listener = l
	return c
}

// lruPrune first removes expired entries, then evicts from the list head.
// Because successful gets move nodes to the tail, the head is the least
// recently used entry.

func lruPrune[K comparable, V any](c *abstractCache[K, V]) int {
	count := 0
	// Remove all expired entries before applying capacity eviction.
	if c.isPruneExpiredActive() {
		for _, key := range c.cacheMap.keysInOrder() {
			co, _ := c.cacheMap.get(key)
			if co.isExpired() {
				c.removeWithoutLock(key)
				count++
			}
		}
	}
	// Evict from the list head until the next insertion can fit.
	for c.capacity > 0 && c.cacheMap.size() >= c.capacity {
		k, ok := c.cacheMap.firstKey()
		if !ok {
			break
		}
		c.removeWithoutLock(k)
		count++
	}
	return count
}
