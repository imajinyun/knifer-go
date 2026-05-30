package cache

// linkedNode is a node in the doubly linked list backing linkedMap.
type linkedNode[K comparable, V any] struct {
	key   K
	value *CacheObj[K, V]
	prev  *linkedNode[K, V]
	next  *linkedNode[K, V]
}

// linkedMap combines a map with a doubly linked list.
// It provides O(1) get/put/remove operations and O(1) movement to the tail for
// LRU. The head stores the oldest entry and the tail stores the newest or most
// recently used entry, depending on the cache strategy.
type linkedMap[K comparable, V any] struct {
	m    map[K]*linkedNode[K, V]
	head *linkedNode[K, V]
	tail *linkedNode[K, V]
}

func newLinkedMap[K comparable, V any](initialCap int) *linkedMap[K, V] {
	if initialCap < 0 {
		initialCap = 0
	}
	return &linkedMap[K, V]{m: make(map[K]*linkedNode[K, V], initialCap)}
}

func (lm *linkedMap[K, V]) size() int { return len(lm.m) }

func (lm *linkedMap[K, V]) get(key K) (*CacheObj[K, V], bool) {
	n, ok := lm.m[key]
	if !ok {
		return nil, false
	}
	return n.value, true
}

// putBack appends a new key/value pair to the list tail.
// Existing keys are replaced in place to preserve their current order.
func (lm *linkedMap[K, V]) putBack(key K, value *CacheObj[K, V]) (old *CacheObj[K, V], existed bool) {
	if n, ok := lm.m[key]; ok {
		old = n.value
		n.value = value
		return old, true
	}
	n := &linkedNode[K, V]{key: key, value: value}
	lm.m[key] = n
	lm.appendNode(n)
	return zeroOf[CacheObj[K, V]](), false
}

// remove deletes key from both the map and the linked list.
func (lm *linkedMap[K, V]) remove(key K) (*CacheObj[K, V], bool) {
	n, ok := lm.m[key]
	if !ok {
		return nil, false
	}
	delete(lm.m, key)
	lm.detach(n)
	return n.value, true
}

// moveToBack moves the node for key to the list tail.
func (lm *linkedMap[K, V]) moveToBack(key K) {
	n, ok := lm.m[key]
	if !ok {
		return
	}
	if lm.tail == n {
		return
	}
	lm.detach(n)
	lm.appendNode(n)
}

// firstKey returns the head key, or the zero value and false when empty.
func (lm *linkedMap[K, V]) firstKey() (K, bool) {
	if lm.head == nil {
		var zero K
		return zero, false
	}
	return lm.head.key, true
}

// keysInOrder returns all keys from head to tail.
func (lm *linkedMap[K, V]) keysInOrder() []K {
	out := make([]K, 0, len(lm.m))
	for n := lm.head; n != nil; n = n.next {
		out = append(out, n.key)
	}
	return out
}

// valuesInOrder returns all cache objects from head to tail.
func (lm *linkedMap[K, V]) valuesInOrder() []*CacheObj[K, V] {
	out := make([]*CacheObj[K, V], 0, len(lm.m))
	for n := lm.head; n != nil; n = n.next {
		out = append(out, n.value)
	}
	return out
}

func (lm *linkedMap[K, V]) clear() {
	lm.m = make(map[K]*linkedNode[K, V])
	lm.head = nil
	lm.tail = nil
}

func (lm *linkedMap[K, V]) appendNode(n *linkedNode[K, V]) {
	n.prev = lm.tail
	n.next = nil
	if lm.tail != nil {
		lm.tail.next = n
	} else {
		lm.head = n
	}
	lm.tail = n
}

func (lm *linkedMap[K, V]) detach(n *linkedNode[K, V]) {
	if n.prev != nil {
		n.prev.next = n.next
	} else {
		lm.head = n.next
	}
	if n.next != nil {
		n.next.prev = n.prev
	} else {
		lm.tail = n.prev
	}
	n.prev = nil
	n.next = nil
}

// zeroOf returns nil for *T.
func zeroOf[T any]() *T { return nil }
