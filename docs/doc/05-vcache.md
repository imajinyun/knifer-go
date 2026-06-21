# vcache Quickstart

`vcache` provides generic cache facades for FIFO, LFU, LRU, timed, weak, and no-cache implementations, with options for capacity, TTL, listeners, and custom clocks.

## Which helper should I use?

| Need | Use | Notes |
| --- | --- | --- |
| Evict by insertion order | `NewFIFO`, `NewFIFOWithOptions`, `NewFIFOWithTimeout` | Good for simple bounded queues where recency does not matter. |
| Evict by least recent use | `NewLRU`, `NewLRUWithOptions`, `NewLRUWithTimeout` | Good default for request/path/object caches with temporal locality. |
| Evict by least frequent use | `NewLFU`, `NewLFUWithOptions`, `NewLFUWithTimeout` | Useful when stable hot keys should survive occasional scans. |
| Store only until TTL | `NewTimed`, `NewTimedWithOptions`, `NewTimedScheduled` | Call `Prune` or schedule pruning to remove expired entries proactively. |
| Disable caching behind an interface | `NewNoCache`, `NewNo` | Useful for tests, feature flags, or bypass modes. |
| Store pointer values with weak-style cleanup | `NewWeak`, `NewWeakWithOptions`, weak finalizer options | Best-effort cleanup only; do not rely on finalizers for correctness. |
| Load values on miss | `GetOrLoad`, `GetOrLoadWith` | Supplier errors are returned and should not be hidden. |
| Observe removals | `WithListener`, `CacheListenerFunc` | Keep listener callbacks fast and non-blocking. |
| Make time deterministic | `WithClock`, `WithTickerFactory`, `WithRunner` | Use provider injection in tests instead of sleeps. |

## Cache correctness checklist

- Choose capacity and TTL from the memory budget and data freshness requirements; unlimited caches can become memory leaks.
- Cache values may be mutable. Store immutable data or defensive copies when callers should not share state through the cache.
- Treat `GetOrLoad` suppliers as side-effecting operations that can fail; propagate errors and avoid expensive duplicate work at a higher layer when needed.
- Keep removal listeners short. Blocking listeners can make cache mutation paths slower or deadlock if they call back into shared application state unsafely.
- Stop scheduled pruning when the cache lifecycle ends if the implementation exposes a cancellation method through the concrete type.
- Do not rely on `WeakCache` finalizers for prompt cleanup; finalizer timing depends on the garbage collector.
- Use deterministic clocks and ticker factories in tests instead of sleeping for real TTLs.

## When not to use vcache

- Use a distributed cache such as Redis when entries must be shared across processes or survive process restarts.
- Use a purpose-built cache when you need admission policies, sharding, metrics, persistence, or stampede protection beyond the facade contract.
- Use plain maps when the data is small, immutable for the request lifetime, and does not need eviction or TTL behavior.

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

## Related packages

- Use `vcron` when cache refresh or cleanup should run on a recurring wall-clock schedule.
- Use `vjob` when cache warming or invalidation can be batched across keys or shards.
- Use `vmap` when ordinary map transformations are enough and eviction policy is unnecessary.

## Benchmarks and trade-offs

Run focused cache tests before and after changing eviction behavior:

```bash
go test ./internal/cache ./vcache
```

Cache performance depends on hit rate, key cardinality, value size, and eviction policy. LRU and LFU update metadata on reads; FIFO is simpler but ignores access locality; timed caches add clock checks and optional pruning work. Benchmark with production-like key distributions before changing policy.

## FAQ

### Which cache should be the default choice?

Use `NewLRU` when recent access is a good predictor of future access. Use FIFO for simple bounded insertion order and LFU when frequent hot keys should be retained.

### Does `GetOrLoad` hide loader errors?

No. Loader errors are returned to the caller. Treat them like backend failures and decide whether to retry, fall back, or surface the error.

### Is `WeakCache` the same as Java weak references?

No. It is a weak-style timed cache with best-effort finalizer cleanup. Do not use finalizer timing as part of application correctness.

### Do expired entries disappear immediately?

Not necessarily. Expired entries are removed when accessed, pruned, or handled by scheduled pruning depending on the cache type and configuration.
