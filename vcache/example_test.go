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

func ExampleNewFIFOWithTimeout() {
	cache := vcache.NewFIFOWithTimeout[string, int](2, time.Minute)
	cache.Put("a", 1)

	fmt.Println(cache.Timeout())
	// Output: 1m0s
}

func ExampleNewFIFOWithOptions() {
	removed := []string{}
	cache := vcache.NewFIFOWithOptions[string, int](
		vcache.WithCapacity[string, int](1),
		vcache.WithListener[string, int](vcache.CacheListenerFunc[string, int](func(key string, value int) {
			removed = append(removed, fmt.Sprintf("%s=%d", key, value))
		})),
	)
	cache.Put("a", 1)
	cache.Put("b", 2)

	fmt.Println(cache.Keys())
	fmt.Println(removed)
	// Output:
	// [b]
	// [a=1]
}

func ExampleNewLFU() {
	cache := vcache.NewLFU[string, int](2)
	cache.Put("a", 1)
	cache.Put("b", 2)

	v, ok := cache.Get("a")
	fmt.Println(v, ok)
	// Output: 1 true
}

func ExampleNewLFUWithTimeout() {
	cache := vcache.NewLFUWithTimeout[string, int](2, time.Minute)
	cache.Put("a", 1)

	fmt.Println(cache.Capacity())
	fmt.Println(cache.Timeout())
	// Output:
	// 2
	// 1m0s
}

func ExampleNewLRUWithTimeout() {
	cache := vcache.NewLRUWithTimeout[string, int](2, time.Minute)

	fmt.Println(cache.Capacity())
	fmt.Println(cache.Timeout())
	// Output:
	// 2
	// 1m0s
}

func ExampleNewTimed() {
	cache := vcache.NewTimed[string, int](time.Minute)
	cache.Put("a", 1)

	fmt.Println(cache.ContainsKey("a"))
	// Output: true
}

func ExampleNewNo() {
	cache := vcache.NewNo[string, int]()
	cache.Put("a", 1)

	fmt.Println(cache.Size(), cache.IsEmpty())
	// Output: 0 true
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

func ExampleNewWeakWithOptions() {
	value := 7
	cache := vcache.NewWeakWithOptions[string, int](
		vcache.WithWeakFinalizerEnabled[string, int](false),
	)
	cache.Put("answer", &value)

	got, ok := cache.Get("answer")
	fmt.Println(*got, ok)
	// Output: 7 true
}

func ExampleCacheListenerFunc() {
	var removed string
	listener := vcache.CacheListenerFunc[string, int](func(key string, value int) {
		removed = fmt.Sprintf("%s=%d", key, value)
	})
	listener.OnRemove("a", 1)

	fmt.Println(removed)
	// Output: a=1
}
