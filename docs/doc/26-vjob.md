# vjob Quickstart

`vjob` provides sliceable task scheduling helpers that split slices, ranges, or map keys into batches and run merge callbacks in order after shards succeed.

## Golden path APIs

The first-choice API set for this facade is kept in sync with `ai-context.json` and the generated tools catalog.

- `NewMapE`
- `NewBatch`
- `NewBatchSingle`
- `NewMap`
- `NewMapKeys`

## Which helper should I use?

| Need | Use | Notes |
| --- | --- | --- |
| Split an index range | `NewSlice`, `NewSliceSingle` | Use for work addressed by integer offsets or pages. |
| Split a typed slice | `NewBatch`, `NewBatchSingle` | Use when callers already have values in memory. |
| Split map keys | `NewMapKeys`, `NewMapE`, `NewMap` | Prefer typed `NewMapKeys`; use `NewMapE` for dynamic input to avoid panics. |
| Execute with embedded options | `Run` | Uses job-provided options when the job carries them. |
| Execute with explicit options | `RunWith`, `Options{BatchSize, MaxConcurrency}` | Use when callers own the scheduling policy. |
| Merge results serially | returned `Merge` callbacks | Merge callbacks run after successful shards and should apply ordered side effects. |
| Control concurrency | `WithMaxConcurrency`, `Options.MaxConcurrency` | Keep concurrency bounded to protect downstream services. |
| Control shard size | `WithBatchSize`, `Options.BatchSize` | Tune for memory, latency, and backend round-trip cost. |

## Job safety checklist

- Pass a non-nil context and honor cancellation in worker functions. `RunWith` returns early when the context is canceled.
- Treat worker functions as concurrent when `MaxConcurrency` is greater than 1. Protect shared mutable state or move mutations into merge callbacks.
- Keep merge callbacks small and deterministic. They run serially, so slow merges reduce total throughput.
- Bound `MaxConcurrency` based on downstream capacity, not CPU count alone. Database, HTTP, and file systems can be overloaded by too many shards.
- Choose `BatchSize` deliberately. Very small batches add scheduling overhead; very large batches reduce cancellation responsiveness and load balancing.
- Prefer `NewMapKeys` or `NewMapE` over `NewMap` for dynamic input. `NewMap` panics on invalid input for compatibility.
- Do not rely on Go map iteration order for deterministic processing. Sort keys before creating a slice-based job when order matters.

## When not to use vjob

- Use `vcron` when work should recur on a wall-clock schedule rather than process one finite collection.
- Use a durable queue or workflow engine when work must survive process crashes, coordinate across nodes, retry persistently, or provide audit trails.
- Use a simple loop when the workload is small, must be strictly sequential, or does not benefit from batching.

## Run range tasks

```go
package main

import (
	"context"
	"fmt"

	"github.com/imajinyun/knifer-go/vjob"
)

func main() {
	job := vjob.NewSlice(func(ctx context.Context, start, end int) (vjob.Merge, error) {
		fmt.Println("run", start, end)
		return func() error {
			fmt.Println("merge", start, end)
			return nil
		}, nil
	}, 10)

	err := vjob.RunWith(context.Background(), job, vjob.Options{
		BatchSize:      3,
		MaxConcurrency: 2,
	})
	if err != nil {
		panic(err)
	}
}
```

## Process typed slices

```go
package main

import (
	"context"
	"fmt"
	"sync"

	"github.com/imajinyun/knifer-go/vjob"
)

func main() {
	values := []int{1, 2, 3, 4}
	var mu sync.Mutex
	total := 0

	job := vjob.NewBatch(func(ctx context.Context, batch []int) (vjob.Merge, error) {
		sum := 0
		for _, v := range batch {
			sum += v
		}
		return func() error {
			mu.Lock()
			defer mu.Unlock()
			total += sum
			return nil
		}, nil
	}, values).WithBatchSize(2).WithMaxConcurrency(2)

	if err := vjob.Run(context.Background(), job); err != nil {
		panic(err)
	}
	fmt.Println(total)
}
```

## Process slice elements one by one

```go
package main

import (
	"context"
	"fmt"

	"github.com/imajinyun/knifer-go/vjob"
)

func main() {
	items := []string{"go", "knifer", "job"}
	job := vjob.NewBatchSingle(func(ctx context.Context, item string) (vjob.Merge, error) {
		upper := len(item)
		return func() error {
			fmt.Println(item, upper)
			return nil
		}, nil
	}, items).WithMaxConcurrency(3)

	if err := vjob.Run(context.Background(), job); err != nil {
		panic(err)
	}
}
```

## Iterate map keys

```go
package main

import (
	"context"
	"fmt"

	"github.com/imajinyun/knifer-go/vjob"
)

func main() {
	scores := map[string]int{"alice": 90, "bob": 80}
	job := vjob.NewMapKeys(func(ctx context.Context, name string) (vjob.Merge, error) {
		return func() error {
			fmt.Println(name, scores[name])
			return nil
		}, nil
	}, scores).WithBatchSize(1)

	if err := vjob.Run(context.Background(), job); err != nil {
		panic(err)
	}
}
```

## Related packages

- Use `vcron` when work should run on recurring wall-clock schedules instead of finite batches.
- Use `vsem` when concurrent workers need weighted permit limits or cancellation-aware throttling.
- Use `vlog` and `verr` when batch failures need structured diagnostics and aggregation.

## Benchmarks and trade-offs

Run focused job tests before changing scheduler behavior:

```bash
go test ./internal/job ./vjob
```

Throughput depends on batch size, merge cost, worker cost, and downstream limits. Increasing concurrency can reduce wall-clock time for I/O-bound shards but can also overload dependencies. Larger batches reduce scheduler overhead but delay cancellation and serial merge progress.

## FAQ

### Are workers run concurrently?

They may be when `MaxConcurrency` is greater than 1. Write worker functions as if they can run in parallel, and use merge callbacks for ordered side effects.

### Are merge callbacks concurrent?

No. Merge callbacks are replayed serially after shards succeed. Keep them short because they become the ordered commit phase.

### What happens on worker errors or panics?

`RunWith` returns an error and does not treat partial success as complete. Design worker side effects so retries are safe or commit through merge callbacks only after success.

### How do I process map keys deterministically?

Extract keys, sort them, then use `NewBatch` or `NewBatchSingle` over the sorted slice. Map iteration order is intentionally unstable.
