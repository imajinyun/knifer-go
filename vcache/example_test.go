package vcache_test

import (
	"fmt"

	"github.com/imajinyun/go-knifer/vcache"
)

func ExampleNewLRU() {
	cache := vcache.NewLRU[string, int](2)
	cache.Put("a", 1)
	cache.Put("b", 2)

	v, ok := cache.Get("a")
	fmt.Println(v, ok)
	// Output: 1 true
}

func ExampleNewFIFO() {
	cache := vcache.NewFIFO[string, int](3)
	cache.Put("x", 10)
	cache.Put("y", 20)

	v, ok := cache.Get("y")
	fmt.Println(v, ok)
	// Output: 20 true
}
