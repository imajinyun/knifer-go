# vsem Quickstart

`vsem` provides counting semaphores with support for weights, context cancellation, non-blocking acquire attempts, and close semantics.

## Which helper should I use?

| Goal | Start with | Notes |
| --- | --- | --- |
| Create a semaphore and panic on invalid capacity | `New` | Convenient for constants and setup code where invalid capacity is a programming error. |
| Create a semaphore and handle invalid capacity | `NewE` | Prefer when capacity comes from configuration or user input. |
| Wait for permits | `Acquire(ctx, weight)` | Blocks until enough permits are available, the context ends, or the semaphore is closed. |
| Try without blocking | `TryAcquire(weight)` | Returns immediately, useful for opportunistic work or fast rejection. |
| Return permits | `Release(weight)` | Always pair successful acquires with release, usually via `defer`. |
| Stop future acquire calls | `Close()` | Wakes waiters and makes future acquires return `ErrClosed`. |
| Inspect state | `Cap()`, `Use()` | Useful for diagnostics and tests; avoid using them as synchronization decisions. |

## Semaphore correctness checklist

- Use `NewE` for configuration-driven capacities so invalid values return `ErrInvalidCapacity` instead of panicking.
- Release only after a successful acquire. Releasing permits that were not acquired can violate concurrency limits.
- Pair `Acquire` with `defer Release` as soon as the acquire succeeds.
- Pass a cancellable context to `Acquire` in request paths so blocked work can exit on timeout or shutdown.
- Use `TryAcquire` when waiting would hold a request, lock, or scheduler thread longer than intended.
- Treat `Close` as a terminal lifecycle event; create a new semaphore instead of reopening a closed one.
- Do not make correctness decisions from `Use()` snapshots because concurrent goroutines can change state immediately after the call.

## Create semaphores and acquire/release permits

```go
package main

import (
	"context"
	"fmt"

	"github.com/imajinyun/go-knifer/vsem"
)

func main() {
	sem := vsem.New(2)
	if err := sem.Acquire(context.Background(), 1); err != nil {
		panic(err)
	}
	fmt.Println(sem.Cap(), sem.Use())
	sem.Release(1)
	fmt.Println(sem.Use())
}
```

## Acquire permits without blocking

```go
package main

import (
	"context"
	"fmt"

	"github.com/imajinyun/go-knifer/vsem"
)

func main() {
	sem := vsem.New(1)
	if err := sem.Acquire(context.Background(), 1); err != nil {
		panic(err)
	}
	fmt.Println(sem.TryAcquire(1))
	sem.Release(1)
	fmt.Println(sem.TryAcquire(1))
	sem.Release(1)
}
```

## Control wait timeouts with context

```go
package main

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/imajinyun/go-knifer/vsem"
)

func main() {
	sem := vsem.New(1)
	if err := sem.Acquire(context.Background(), 1); err != nil {
		panic(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
	defer cancel()
	err := sem.Acquire(ctx, 1)
	fmt.Println(errors.Is(err, context.DeadlineExceeded))
	sem.Release(1)
}
```

## Check errors and close semaphores

```go
package main

import (
	"context"
	"errors"
	"fmt"

	"github.com/imajinyun/go-knifer/vsem"
)

func main() {
	sem, err := vsem.NewE(0)
	fmt.Println(sem == nil, errors.Is(err, vsem.ErrInvalidCapacity))

	active := vsem.New(1)
	active.Close()
	err = active.Acquire(context.Background(), 1)
	fmt.Println(errors.Is(err, vsem.ErrClosed))
}
```

## When not to use vsem

- Use a buffered channel for very small local concurrency limits that do not need weights, close errors, or introspection.
- Use a worker pool when tasks should be queued, retried, drained, or supervised by worker lifecycle logic.
- Use `sync.Mutex` or `sync.RWMutex` when protecting shared memory rather than limiting concurrent work.
- Use distributed locks or rate limiters when concurrency must be coordinated across processes.

## Related packages

- Use `vjob` when permit-limited work is naturally expressed as slice, map-key, or range batches.
- Use `vcron` when concurrency-limited work should be scheduled on recurring wall-clock intervals.
- Use `verr` when worker failures need aggregation or panic recovery.

## Benchmarks and trade-offs

- Weighted semaphores are more expressive than buffered channels but require careful release accounting.
- Context-aware blocking prevents goroutine leaks in shutdown and timeout paths, with a small amount of coordination overhead.
- `TryAcquire` avoids blocking overhead but can reject work during short bursts that would have succeeded after waiting.
- `Use` and `Cap` are useful observability hooks, but reading them in hot loops can add lock contention.
- Closing a semaphore is a clear shutdown signal; it is not a pause/resume mechanism.

## FAQ

### Should I call `Release` after `Acquire` returns an error?

No. Release only permits that were successfully acquired. If `Acquire` returns context cancellation or `ErrClosed`, no permits were granted.

### What happens to waiters when `Close` is called?

Blocked and future acquire calls fail with `ErrClosed`. Existing holders should still release according to their own cleanup flow.

### When should I use weighted permits?

Use weights when tasks consume different amounts of a limited resource, such as memory slots, API quota units, or expensive worker capacity.

### Is `TryAcquire` fair?

It is an immediate capacity check, not a fairness policy. Use blocking `Acquire` with context when work should wait its turn.
