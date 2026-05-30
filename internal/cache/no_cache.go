package cache

import "time"

// NoCache implements Cache without storing any values.
type NoCache[K comparable, V any] struct{}

// NewNoCache creates a no-op cache.
func NewNoCache[K comparable, V any]() *NoCache[K, V] { return &NoCache[K, V]{} }

func (NoCache[K, V]) Capacity() int                                        { return 0 }
func (NoCache[K, V]) Timeout() time.Duration                               { return 0 }
func (NoCache[K, V]) Put(key K, value V)                                   {}
func (NoCache[K, V]) PutWithTimeout(key K, value V, timeout time.Duration) {}

func (NoCache[K, V]) Get(key K) (V, bool) {
	var zero V
	return zero, false
}

func (NoCache[K, V]) GetWithUpdate(key K, _ bool) (V, bool) {
	var zero V
	return zero, false
}

func (NoCache[K, V]) GetOrLoad(key K, supplier Supplier[V]) (V, error) {
	if supplier == nil {
		var zero V
		return zero, nil
	}
	return supplier()
}

func (n NoCache[K, V]) GetOrLoadWith(key K, _ bool, _ time.Duration, supplier Supplier[V]) (V, error) {
	return n.GetOrLoad(key, supplier)
}

func (NoCache[K, V]) Remove(key K)           {}
func (NoCache[K, V]) ContainsKey(key K) bool { return false }
func (NoCache[K, V]) Size() int              { return 0 }
func (NoCache[K, V]) IsEmpty() bool          { return true }
func (NoCache[K, V]) IsFull() bool           { return false }
func (NoCache[K, V]) Prune() int             { return 0 }
func (NoCache[K, V]) Clear()                 {}
func (NoCache[K, V]) Keys() []K              { return nil }
func (NoCache[K, V]) Values() []V            { return nil }

func (n NoCache[K, V]) SetListener(_ CacheListener[K, V]) Cache[K, V] { return n }
func (NoCache[K, V]) HitCount() int64                                 { return 0 }
func (NoCache[K, V]) MissCount() int64                                { return 0 }
