package vcache_test

import (
	"testing"

	"github.com/imajinyun/go-knifer/vcache"
)

func TestFacadeNoCache(t *testing.T) {
	c := vcache.NewNoCache[string, int]()
	c.Put("a", 1)
	if _, ok := c.Get("a"); ok {
		t.Fatal("expected NoCache to store nothing")
	}
}

func TestFacadeNo(t *testing.T) {
	c := vcache.NewNo[string, int]()
	c.Put("a", 1)
	if _, ok := c.Get("a"); ok {
		t.Fatal("expected NewNo cache to store nothing")
	}
}
