package cache

import (
	"sync"
	"sync/atomic"
	"time"
)

// pruneStrategy is the eviction strategy provided by each concrete cache.
// It returns the number of entries removed while abstractCache.mu is held.
type pruneStrategy[K comparable, V any] func(c *abstractCache[K, V]) int

// abstractCache contains the shared implementation for cache variants, similar
// to the utility cache AbstractCache and ReentrantCache.
//
// A single mutex protects both the map/list structure and metadata updates.
// Reads also take the lock because a successful get may refresh the last access
// time, increment the hit counter on the entry, and move the node for LRU.
type abstractCache[K comparable, V any] struct {
	mu        sync.Mutex
	cacheMap  *linkedMap[K, V]
	capacity  int
	timeout   time.Duration
	pruneFn   pruneStrategy[K, V]
	listener  CacheListener[K, V]
	hitCount  int64
	missCount int64
	clock     func() time.Time
	ticker    TickerFactory
	runner    func(func())

	// existCustomTimeout records whether any entry uses a non-default TTL.
	existCustomTimeout bool

	// moveToBackOnGet moves a node to the list tail after a successful get.
	moveToBackOnGet bool

	// keyLocks serializes GetOrLoad calls per key to prevent duplicate loading.
	keyLocks sync.Map
}

func (c *abstractCache[K, V]) init(capacity int, timeout time.Duration, prune pruneStrategy[K, V]) {
	c.capacity = capacity
	c.timeout = timeout
	c.pruneFn = prune
	c.cacheMap = newLinkedMap[K, V](capacity)
	c.clock = time.Now
	c.ticker = newTicker
	c.runner = defaultRunner
}

func (c *abstractCache[K, V]) Capacity() int          { return c.capacity }
func (c *abstractCache[K, V]) Timeout() time.Duration { return c.timeout }
func (c *abstractCache[K, V]) HitCount() int64        { return atomic.LoadInt64(&c.hitCount) }
func (c *abstractCache[K, V]) MissCount() int64       { return atomic.LoadInt64(&c.missCount) }

func (c *abstractCache[K, V]) setClock(clock func() time.Time) {
	if clock != nil {
		c.clock = clock
	}
}

func (c *abstractCache[K, V]) now() time.Time {
	if c.clock != nil {
		return c.clock()
	}
	return time.Now()
}

func (c *abstractCache[K, V]) setTickerFactory(factory TickerFactory) {
	if factory != nil {
		c.ticker = factory
	}
}

func (c *abstractCache[K, V]) newTicker(delay time.Duration) (<-chan time.Time, Ticker) {
	if c.ticker != nil {
		return c.ticker(delay)
	}
	return newTicker(delay)
}

func (c *abstractCache[K, V]) setRunner(runner func(func())) {
	if runner != nil {
		c.runner = runner
	}
}

func (c *abstractCache[K, V]) run(fn func()) {
	if c.runner != nil {
		c.runner(fn)
		return
	}
	defaultRunner(fn)
}

// IsFull reports whether the cache has reached its configured capacity.
func (c *abstractCache[K, V]) IsFull() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.isFullLocked()
}

func (c *abstractCache[K, V]) isFullLocked() bool {
	return c.capacity > 0 && c.cacheMap.size() >= c.capacity
}

// isPruneExpiredActive reports whether expiration checks are needed.
func (c *abstractCache[K, V]) isPruneExpiredActive() bool {
	return c.timeout > 0 || c.existCustomTimeout
}

// Size returns the number of entries currently stored in the cache.
func (c *abstractCache[K, V]) Size() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.cacheMap.size()
}

// IsEmpty reports whether the cache contains no entries.
func (c *abstractCache[K, V]) IsEmpty() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.cacheMap.size() == 0
}

// Put stores an entry using the cache's default expiration duration.
func (c *abstractCache[K, V]) Put(key K, value V) {
	c.PutWithTimeout(key, value, c.timeout)
}

// PutWithTimeout stores an entry with a custom expiration duration.
func (c *abstractCache[K, V]) PutWithTimeout(key K, value V, timeout time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.putLocked(key, value, timeout)
}

func (c *abstractCache[K, V]) putLocked(key K, value V, timeout time.Duration) {
	co := newCacheObj(key, value, timeout, c.now())
	if timeout > 0 && timeout != c.timeout {
		c.existCustomTimeout = true
	}
	if old, ok := c.cacheMap.get(key); ok {
		// Replacing an existing entry must not trigger capacity eviction.
		c.cacheMap.putBack(key, co)
		c.notifyRemove(old.key, old.value)
		return
	}
	if c.isFullLocked() {
		c.pruneFn(c)
	}
	c.cacheMap.putBack(key, co)
}

// Get returns a cached value and refreshes access metadata on hit.
// Missing or expired entries return the zero value and false.
func (c *abstractCache[K, V]) Get(key K) (V, bool) {
	return c.GetWithUpdate(key, true)
}

// GetWithUpdate returns a cached value and optionally refreshes last access time.
func (c *abstractCache[K, V]) GetWithUpdate(key K, updateLastAccess bool) (V, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.getLocked(key, updateLastAccess)
}

func (c *abstractCache[K, V]) getLocked(key K, updateLastAccess bool) (V, bool) {
	var zero V
	co, ok := c.cacheMap.get(key)
	if !ok {
		atomic.AddInt64(&c.missCount, 1)
		return zero, false
	}
	if co.isExpired(c.now()) {
		// Remove expired entries immediately so future lookups do not see them.
		c.cacheMap.remove(key)
		c.notifyRemove(co.key, co.value)
		atomic.AddInt64(&c.missCount, 1)
		return zero, false
	}
	v := co.get(updateLastAccess, c.now())
	atomic.AddInt64(&c.hitCount, 1)
	c.afterGet(key)
	return v, true
}

// afterGet is a hook used by LRU to move a hit entry to the list tail.
func (c *abstractCache[K, V]) afterGet(key K) {
	if c.moveToBackOnGet {
		c.cacheMap.moveToBack(key)
	}
}

// GetOrLoad calls supplier on cache miss and stores the generated value.
func (c *abstractCache[K, V]) GetOrLoad(key K, supplier Supplier[V]) (V, error) {
	return c.GetOrLoadWith(key, true, c.timeout, supplier)
}

// GetOrLoadWith calls supplier on cache miss and stores the generated value.
// The caller can control access-time refresh and the TTL used for the loaded value.
func (c *abstractCache[K, V]) GetOrLoadWith(key K, updateLastAccess bool, timeout time.Duration, supplier Supplier[V]) (V, error) {
	if v, ok := c.GetWithUpdate(key, updateLastAccess); ok {
		return v, nil
	}
	if supplier == nil {
		var zero V
		return zero, nil
	}
	// Double-check after acquiring the per-key lock; another goroutine may have
	// populated the same key while this goroutine was waiting.
	lockAny, _ := c.keyLocks.LoadOrStore(key, &sync.Mutex{})
	lock := lockAny.(*sync.Mutex)
	lock.Lock()
	defer func() {
		lock.Unlock()
		c.keyLocks.Delete(key)
	}()
	if v, ok := c.GetWithUpdate(key, updateLastAccess); ok {
		return v, nil
	}
	v, err := supplier()
	if err != nil {
		return v, err
	}
	c.PutWithTimeout(key, v, timeout)
	return v, nil
}

// ContainsKey reports whether key exists and removes it if the entry has expired.
func (c *abstractCache[K, V]) ContainsKey(key K) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	co, ok := c.cacheMap.get(key)
	if !ok {
		return false
	}
	if co.isExpired(c.now()) {
		c.cacheMap.remove(key)
		c.notifyRemove(co.key, co.value)
		return false
	}
	return true
}

// Remove deletes one key and notifies the removal listener when present.
func (c *abstractCache[K, V]) Remove(key K) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if old, ok := c.cacheMap.remove(key); ok {
		c.notifyRemove(old.key, old.value)
	}
}

// Clear removes all entries and notifies the listener for each removed value.
func (c *abstractCache[K, V]) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	for _, co := range c.cacheMap.valuesInOrder() {
		c.notifyRemove(co.key, co.value)
	}
	c.cacheMap.clear()
}

// Prune runs the configured eviction strategy and returns the removed count.
func (c *abstractCache[K, V]) Prune() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.pruneFn(c)
}

// Keys returns a snapshot of all keys in list order.
func (c *abstractCache[K, V]) Keys() []K {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.cacheMap.keysInOrder()
}

// Values returns a snapshot of all non-expired values in list order.
func (c *abstractCache[K, V]) Values() []V {
	c.mu.Lock()
	defer c.mu.Unlock()
	out := make([]V, 0, c.cacheMap.size())
	for _, co := range c.cacheMap.valuesInOrder() {
		if !co.isExpired(c.now()) {
			out = append(out, co.value)
		}
	}
	return out
}

func (c *abstractCache[K, V]) notifyRemove(key K, value V) {
	if c.listener != nil {
		c.listener.OnRemove(key, value)
	}
}

// removeWithoutLock removes a key and notifies the listener.
// Callers must already hold abstractCache.mu.
func (c *abstractCache[K, V]) removeWithoutLock(key K) {
	if old, ok := c.cacheMap.remove(key); ok {
		c.notifyRemove(old.key, old.value)
	}
}
