package vcache_test

import (
	"testing"
	"time"

	"github.com/imajinyun/go-knifer/vcache"
)

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

func TestFacadeLRUWithTimeout(t *testing.T) {
	c := vcache.NewLRUWithTimeout[string, int](2, time.Second)
	c.Put("a", 1)
	c.Put("b", 2)
	if c.Capacity() != 2 || c.Timeout() != time.Second {
		t.Fatalf("LRUWithTimeout capacity=%d timeout=%s", c.Capacity(), c.Timeout())
	}
}

func TestFacadeLFUCache(t *testing.T) {
	c := vcache.NewLFU[string, int](2)
	c.Put("a", 1)
	c.Put("b", 2)
	c.Put("c", 3) // evicts least frequently used
	if c.Capacity() != 2 {
		t.Fatalf("LFU capacity=%d", c.Capacity())
	}
}

func TestFacadeLFUWithTimeout(t *testing.T) {
	c := vcache.NewLFUWithTimeout[string, int](2, time.Second)
	if c.Capacity() != 2 || c.Timeout() != time.Second {
		t.Fatalf("LFUWithTimeout capacity=%d timeout=%s", c.Capacity(), c.Timeout())
	}
}
