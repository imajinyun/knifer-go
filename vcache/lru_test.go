package vcache_test

import (
	"testing"

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
