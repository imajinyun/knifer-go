package vcache_test

import (
	"testing"
	"time"

	"github.com/imajinyun/go-knifer/vcache"
)

func TestFacadeFIFOCache(t *testing.T) {
	c := vcache.NewFIFO[string, int](2)
	c.Put("a", 1)
	c.Put("b", 2)
	c.Put("c", 3) // evicts "a"

	if _, ok := c.Get("a"); ok {
		t.Fatal("expected 'a' to be evicted from FIFO cache")
	}
	if v, ok := c.Get("b"); !ok || v != 2 {
		t.Fatalf("expected b=2, got %v, ok=%v", v, ok)
	}
}

func TestFacadeFIFOWithTimeout(t *testing.T) {
	c := vcache.NewFIFOWithTimeout[string, int](2, time.Second)
	c.Put("a", 1)
	c.Put("b", 2)
	c.Put("c", 3) // evicts "a"

	if _, ok := c.Get("a"); ok {
		t.Fatal("expected 'a' to be evicted from FIFO cache with timeout")
	}
	if c.Capacity() != 2 || c.Timeout() != time.Second {
		t.Fatalf("FIFOWithTimeout capacity=%d timeout=%s", c.Capacity(), c.Timeout())
	}
}

func TestFacadeCacheListener(t *testing.T) {
	var removedKey string
	c := vcache.NewFIFO[string, int](1)
	c.SetListener(vcache.CacheListenerFunc[string, int](func(key string, value int) {
		removedKey = key
	}))
	c.Put("a", 1)
	c.Put("b", 2) // evicts "a"

	if removedKey != "a" {
		t.Fatalf("expected listener to receive 'a', got %q", removedKey)
	}
}

func TestFacadeCacheOptions(t *testing.T) {
	var removedKey string
	fifo := vcache.NewFIFOWithOptions[string, int](
		vcache.WithCapacity[string, int](1),
		vcache.WithTimeout[string, int](time.Second),
		vcache.WithListener[string, int](vcache.CacheListenerFunc[string, int](func(key string, value int) {
			removedKey = key
		})),
	)
	if fifo.Capacity() != 1 || fifo.Timeout() != time.Second {
		t.Fatalf("FIFO options not applied: capacity=%d timeout=%s", fifo.Capacity(), fifo.Timeout())
	}
	fifo.Put("a", 1)
	fifo.Put("b", 2)
	if removedKey != "a" {
		t.Fatalf("expected listener to receive 'a', got %q", removedKey)
	}

	lfu := vcache.NewLFUWithOptions[string, int](vcache.WithCapacity[string, int](2), vcache.WithTimeout[string, int](time.Second))
	if lfu.Capacity() != 2 || lfu.Timeout() != time.Second {
		t.Fatalf("LFU options not applied: capacity=%d timeout=%s", lfu.Capacity(), lfu.Timeout())
	}
	lru := vcache.NewLRUWithOptions[string, int](vcache.WithCapacity[string, int](2), vcache.WithTimeout[string, int](time.Second))
	if lru.Capacity() != 2 || lru.Timeout() != time.Second {
		t.Fatalf("LRU options not applied: capacity=%d timeout=%s", lru.Capacity(), lru.Timeout())
	}
	timed := vcache.NewTimedWithOptions[string, int](vcache.WithTimeout[string, int](time.Second))
	if timed.Capacity() != 0 || timed.Timeout() != time.Second {
		t.Fatalf("Timed options not applied: capacity=%d timeout=%s", timed.Capacity(), timed.Timeout())
	}
}

func TestFacadeCacheWithWeakOptions(t *testing.T) {
	finalizerFunc := vcache.WithWeakFinalizerFunc[string, int](func(v *int, done func(*int)) {
		// NOTE: done(v) cannot be called synchronously during Put
		// because it re-acquires the WeakCache lock.
	})
	if finalizerFunc == nil {
		t.Fatal("WithWeakFinalizerFunc returned nil")
	}

	enabledOpt := vcache.WithWeakFinalizerEnabled[string, int](true)
	if enabledOpt == nil {
		t.Fatal("WithWeakFinalizerEnabled returned nil")
	}

	v := 7
	weak := vcache.NewWeakWithOptions[string, int](
		vcache.WithTimeout[string, *int](time.Hour),
		vcache.WithWeakFinalizerFunc[string, int](func(v *int, done func(*int)) {}),
		vcache.WithWeakFinalizerEnabled[string, int](true),
	)
	weak.Put("x", &v)
	if got, ok := weak.Get("x"); !ok || *got != 7 {
		t.Fatalf("weak cache with options: got %v, ok=%v", got, ok)
	}
}

func TestFacadeCacheWithRunner(t *testing.T) {
	runnerOpt := vcache.WithRunner[string, int](func(fn func()) {
		fn()
	})
	if runnerOpt == nil {
		t.Fatal("WithRunner returned nil")
	}
}
