package cache

import "time"

// CacheListener receives callbacks when entries are removed from a cache.
type CacheListener[K comparable, V any] interface {
	OnRemove(key K, value V)
}

// CacheListenerFunc adapts a function to CacheListener.
type CacheListenerFunc[K comparable, V any] func(key K, value V)

// OnRemove implements CacheListener.
func (f CacheListenerFunc[K, V]) OnRemove(key K, value V) { f(key, value) }

// Supplier creates a value when GetOrLoad observes a cache miss.
type Supplier[V any] func() (V, error)

// Cache defines the common cache operations, similar to hutool-cache Cache.
type Cache[K comparable, V any] interface {
	// Capacity returns the maximum number of entries; 0 means unlimited.
	Capacity() int
	// Timeout returns the default expiration duration; 0 means no expiration.
	Timeout() time.Duration
	// Put stores an entry using the default expiration duration.
	Put(key K, value V)
	// PutWithTimeout stores an entry using the specified expiration duration.
	PutWithTimeout(key K, value V, timeout time.Duration)
	// Get returns a value and refreshes access metadata on hit.
	Get(key K) (V, bool)
	// GetWithUpdate returns a value and optionally refreshes last access time.
	GetWithUpdate(key K, updateLastAccess bool) (V, bool)
	// GetOrLoad calls supplier on miss and stores the generated value.
	GetOrLoad(key K, supplier Supplier[V]) (V, error)
	// GetOrLoadWith calls supplier on miss and controls refresh and expiration.
	GetOrLoadWith(key K, updateLastAccess bool, timeout time.Duration, supplier Supplier[V]) (V, error)
	// Remove deletes one key.
	Remove(key K)
	// ContainsKey reports whether key exists and prunes it if expired.
	ContainsKey(key K) bool
	// Size returns the current number of entries.
	Size() int
	// IsEmpty reports whether the cache contains no entries.
	IsEmpty() bool
	// IsFull reports whether the cache reached its capacity.
	IsFull() bool
	// Prune removes expired or evicted entries and returns the removed count.
	Prune() int
	// Clear removes all entries.
	Clear()
	// Keys returns a snapshot of all keys.
	Keys() []K
	// Values returns a snapshot of all non-expired values.
	Values() []V
	// SetListener sets the removal listener and returns the cache for chaining.
	SetListener(listener CacheListener[K, V]) Cache[K, V]
	// HitCount returns the number of successful lookups.
	HitCount() int64
	// MissCount returns the number of missed or expired lookups.
	MissCount() int64
}
