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

func TestFacadeLRUCache(t *testing.T) {
	c := vcache.NewLRU[string, int](2)
	c.Put("a", 1)
	c.Put("b", 2)
	c.Get("a")    // touch "a"
	c.Put("c", 3) // evicts "b" (least recently used)

	if _, ok := c.Get("b"); ok {
		t.Fatal("expected 'b' to be evicted from LRU cache")
	}
}

func TestFacadeTimedCache(t *testing.T) {
	c := vcache.NewTimed[string, int](50 * time.Millisecond)
	c.Put("x", 10)
	if v, ok := c.Get("x"); !ok || v != 10 {
		t.Fatalf("expected x=10, got %v, ok=%v", v, ok)
	}
	time.Sleep(100 * time.Millisecond)
	if _, ok := c.Get("x"); ok {
		t.Fatal("expected 'x' to expire from timed cache")
	}
}

func TestFacadeNoCache(t *testing.T) {
	c := vcache.NewNoCache[string, int]()
	c.Put("a", 1)
	if _, ok := c.Get("a"); ok {
		t.Fatal("expected NoCache to store nothing")
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
