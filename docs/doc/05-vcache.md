# vcache Quickstart

`vcache` provides generic cache facades for FIFO, LFU, LRU, timed, weak, and no-cache implementations, with options for capacity, TTL, listeners, and custom clocks.

## LRU cache

```go
package main

import (
	"fmt"

	"github.com/imajinyun/go-knifer/vcache"
)

func main() {
	c := vcache.NewLRU[string, int](2)
	c.Put("a", 1)
	c.Put("b", 2)
	c.Get("a")
	c.Put("c", 3)

	_, ok := c.Get("b")
	fmt.Println("b exists:", ok)
}
```

## FIFO cache and removal listeners

```go
package main

import (
	"fmt"

	"github.com/imajinyun/go-knifer/vcache"
)

func main() {
	removed := make([]string, 0)
	c := vcache.NewFIFOWithOptions[string, int](
		vcache.WithCapacity[string, int](1),
		vcache.WithListener[string, int](vcache.CacheListenerFunc[string, int](func(key string, value int) {
			removed = append(removed, key)
		})),
	)

	c.Put("first", 1)
	c.Put("second", 2)
	fmt.Println(removed)
}
```

## Timed cache and expiration cleanup

```go
package main

import (
	"fmt"
	"time"

	"github.com/imajinyun/go-knifer/vcache"
)

func main() {
	c := vcache.NewTimed[string, string](50 * time.Millisecond)
	c.Put("token", "abc")
	time.Sleep(80 * time.Millisecond)

	fmt.Println("pruned:", c.Prune())
	_, ok := c.Get("token")
	fmt.Println("token exists:", ok)
}
```

## Load cache values on demand

```go
package main

import (
	"fmt"

	"github.com/imajinyun/go-knifer/vcache"
)

func main() {
	c := vcache.NewLFU[string, string](10)
	value, err := c.GetOrLoad("profile", func() (string, error) {
		return "loaded", nil
	})
	if err != nil {
		panic(err)
	}
	fmt.Println(value)
}
```
