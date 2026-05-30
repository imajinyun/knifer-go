package cache

import (
	"sync"
	"time"
)

// CacheObj stores one cached entry and its access metadata.
type CacheObj[K comparable, V any] struct {
	key         K
	value       V
	ttl         time.Duration // 0 means the entry never expires.
	lastAccess  int64         // Unix timestamp in nanoseconds.
	accessCount int64         // Number of successful accesses.
	mu          sync.Mutex
}

// newCacheObj creates a CacheObj with its last access time set to now.
func newCacheObj[K comparable, V any](key K, value V, ttl time.Duration) *CacheObj[K, V] {
	return &CacheObj[K, V]{
		key:        key,
		value:      value,
		ttl:        ttl,
		lastAccess: time.Now().UnixNano(),
	}
}

// Key returns the entry key.
func (c *CacheObj[K, V]) Key() K { return c.key }

// Value returns the entry value.
func (c *CacheObj[K, V]) Value() V { return c.value }

// TTL returns the entry expiration duration.
func (c *CacheObj[K, V]) TTL() time.Duration { return c.ttl }

// LastAccess returns the last successful access time.
func (c *CacheObj[K, V]) LastAccess() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	return time.Unix(0, c.lastAccess)
}

// AccessCount returns the number of successful accesses.
func (c *CacheObj[K, V]) AccessCount() int64 {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.accessCount
}

// ExpiredTime returns the expiration time, or false when the entry never expires.
func (c *CacheObj[K, V]) ExpiredTime() (time.Time, bool) {
	if c.ttl <= 0 {
		return time.Time{}, false
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	return time.Unix(0, c.lastAccess).Add(c.ttl), true
}

// isExpired reports whether the entry has expired relative to last access time.
func (c *CacheObj[K, V]) isExpired() bool {
	if c.ttl <= 0 {
		return false
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	return time.Now().UnixNano()-c.lastAccess > int64(c.ttl)
}

// get returns the value and updates access count and, optionally, last access time.
func (c *CacheObj[K, V]) get(updateLastAccess bool) V {
	c.mu.Lock()
	defer c.mu.Unlock()
	if updateLastAccess {
		c.lastAccess = time.Now().UnixNano()
	}
	c.accessCount++
	return c.value
}

// addAccessCount adjusts access count for LFU aging and returns the new count.
func (c *CacheObj[K, V]) addAccessCount(delta int64) int64 {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.accessCount += delta
	return c.accessCount
}
