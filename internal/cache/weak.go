package cache

import (
	"runtime"
	"sync"
	"time"
)

// WeakCache is a weak-reference-like cache for pointer values.
//
// Go does not provide Java-style WeakReference. This implementation uses
// runtime.SetFinalizer to approximate weak-reference behavior: when all strong
// references to a cached pointer disappear, a later GC cycle may run the
// finalizer and remove the corresponding entry.
//
// Because finalizer scheduling is intentionally non-deterministic in Go, callers
// should treat GC-based cleanup as eventual cleanup. TTL checks and explicit
// Prune/Remove/Clear remain deterministic.
type WeakCache[K comparable, V any] struct {
	mu       sync.Mutex
	entries  map[K]*weakEntry[V]
	timeout  time.Duration
	listener CacheListener[K, *V]
	hits     int64
	misses   int64
}

type weakEntry[V any] struct {
	ref        *V
	lastAccess int64
	ttl        time.Duration
}

// NewWeakCache creates a weak-reference-like cache with timeout as default TTL.
// A zero timeout means entries do not expire by time.
func NewWeakCache[K comparable, V any](timeout time.Duration) *WeakCache[K, V] {
	return &WeakCache[K, V]{
		entries: make(map[K]*weakEntry[V]),
		timeout: timeout,
	}
}

// SetListener sets the removal listener and returns the cache for chaining.
func (c *WeakCache[K, V]) SetListener(l CacheListener[K, *V]) *WeakCache[K, V] {
	c.listener = l
	return c
}

// Put stores a pointer value using the default timeout.
func (c *WeakCache[K, V]) Put(key K, value *V) {
	c.PutWithTimeout(key, value, c.timeout)
}

// PutWithTimeout stores a pointer value using a custom timeout.
func (c *WeakCache[K, V]) PutWithTimeout(key K, value *V, timeout time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if old, ok := c.entries[key]; ok {
		c.notifyRemove(key, old.ref)
	}
	if value == nil {
		delete(c.entries, key)
		return
	}
	c.entries[key] = &weakEntry[V]{
		ref:        value,
		lastAccess: time.Now().UnixNano(),
		ttl:        timeout,
	}
	// Use a finalizer to remove the entry after the pointed value is collected.
	keyCopy := key
	cache := c
	runtime.SetFinalizer(value, func(v *V) {
		cache.removeIfRefIs(keyCopy, v)
	})
}

// Get returns a cached pointer, or nil and false when missing or expired.
func (c *WeakCache[K, V]) Get(key K) (*V, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	e, ok := c.entries[key]
	if !ok {
		c.misses++
		return nil, false
	}
	if e.ttl > 0 && time.Now().UnixNano()-e.lastAccess > int64(e.ttl) {
		delete(c.entries, key)
		c.notifyRemove(key, e.ref)
		c.misses++
		return nil, false
	}
	e.lastAccess = time.Now().UnixNano()
	c.hits++
	return e.ref, true
}

// Remove deletes one key and notifies the removal listener when present.
func (c *WeakCache[K, V]) Remove(key K) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if e, ok := c.entries[key]; ok {
		delete(c.entries, key)
		c.notifyRemove(key, e.ref)
	}
}

// Size returns the number of entries still tracked internally.
// Values already collected by GC may still be counted until their finalizers run.
func (c *WeakCache[K, V]) Size() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return len(c.entries)
}

// Clear removes all entries and notifies the listener for each value.
func (c *WeakCache[K, V]) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	for k, e := range c.entries {
		c.notifyRemove(k, e.ref)
	}
	c.entries = make(map[K]*weakEntry[V])
}

// Prune removes expired entries and returns the removed count.
func (c *WeakCache[K, V]) Prune() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	count := 0
	now := time.Now().UnixNano()
	for k, e := range c.entries {
		if e.ttl > 0 && now-e.lastAccess > int64(e.ttl) {
			delete(c.entries, k)
			c.notifyRemove(k, e.ref)
			count++
		}
	}
	return count
}

// HitCount returns the number of successful lookups.
func (c *WeakCache[K, V]) HitCount() int64 {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.hits
}

// MissCount returns the number of missed or expired lookups.
func (c *WeakCache[K, V]) MissCount() int64 {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.misses
}

// removeIfRefIs is called by the finalizer and removes key only when it still
// points to the same value. This avoids deleting a newer value stored under the
// same key after the finalizer was registered.
func (c *WeakCache[K, V]) removeIfRefIs(key K, ref *V) {
	c.mu.Lock()
	defer c.mu.Unlock()
	e, ok := c.entries[key]
	if !ok {
		return
	}
	if e.ref == ref {
		delete(c.entries, key)
		c.notifyRemove(key, ref)
	}
}

func (c *WeakCache[K, V]) notifyRemove(key K, value *V) {
	if c.listener != nil {
		c.listener.OnRemove(key, value)
	}
}
