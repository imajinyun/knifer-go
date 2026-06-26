package vcache_test

import (
	"fmt"
	"time"

	"github.com/imajinyun/knifer-go/vcache"
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

func ExampleNewLFU() {
	cache := vcache.NewLFU[string, int](2)
	cache.Put("a", 1)
	cache.Put("b", 2)

	v, ok := cache.Get("a")
	fmt.Println(v, ok)
	// Output: 1 true
}

func ExampleNewNoCache() {
	cache := vcache.NewNoCache[string, int]()
	cache.Put("a", 1)

	_, ok := cache.Get("a")
	fmt.Println(ok)
	// Output: false
}

func ExampleNewTimedWithOptions() {
	base := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	now := base

	cache := vcache.NewTimedWithOptions[string, int](
		vcache.WithTimeout[string, int](time.Second),
		vcache.WithClock[string, int](func() time.Time { return now }),
	)
	cache.Put("a", 1)

	_, before := cache.Get("a")
	now = base.Add(2 * time.Second)
	_, after := cache.Get("a")

	fmt.Println(before, after)
	// Output: true false
}
