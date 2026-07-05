# vcron Quickstart

`vcron` provides cron expression parsing and task scheduling facades, with support for the default scheduler, local schedulers, second-level matching, custom IDs, clocks, and executors.

## Golden path APIs

The first-choice API set for this facade is kept in sync with `ai-context.json` and the generated tools catalog.

- `CronRestart`
- `ConfigureDefaultScheduler`
- `NewConfig`
- `MustNewPattern`
- `CronLaunchingCount`

## Which helper should I use?

| Need | Use | Notes |
| --- | --- | --- |
| Parse cron expressions | `NewPattern`, `NewPatternWithOptions`, `MustNewPattern` | Prefer error-returning constructors for dynamic input. |
| Validate config without scheduling | `NewConfig`, `NewConfigWithOptions`, `WithConfigLocation`, `WithConfigMatchSecond` | Useful for setup and tests. |
| Create an isolated scheduler | `NewScheduler`, `NewSchedulerWithOptions` | Prefer local schedulers for libraries and tests. |
| Use package-level scheduling | `CronSchedule*`, `CronRemove*`, `CronStart`, `CronStop`, `CronShutdown` | Convenient for applications with one global scheduler. |
| Customize execution | `WithExecutor`, `WithRunner`, `WithClock`, `WithSleeper` | Inject deterministic providers for tests and bounded executors for production. |
| Use explicit task IDs | `WithIDGenerator`, `WithIDRandomReader`, `ScheduleWithID` | Stable IDs simplify updates, removal, and observability. |
| Change time zone or seconds mode | `WithLocation`, `WithMatchSecond`, `CronSetMatchSecondE` | Some config becomes immutable after scheduler start. |
| Observe lifecycle | `TaskListener`, `SimpleTaskListener`, running/launching counters | Listeners should be fast and panic-safe. |

## Scheduler safety checklist

- Prefer local schedulers in libraries and tests so package-level state does not leak across callers.
- Always stop or shut down schedulers during application shutdown and tests. Use `CronShutdown(ctx)` or scheduler `Shutdown` when running tasks need to drain.
- Pass bounded contexts to shutdown paths; shutdown waits for running tasks but does not forcibly cancel work already executing.
- Keep task functions idempotent where possible. Cron schedules can overlap if task runtime exceeds the interval unless the task coordinates its own concurrency.
- Keep listeners and executors lightweight. Long-running listeners or unbounded executor goroutines can create backpressure or resource leaks.
- Use `NewPattern` for user-provided expressions and reserve `MustNewPattern` for constants that should fail during tests or startup.
- Be explicit about time zones and whether seconds are part of the expression; mismatched assumptions are a common production scheduling bug.

## When not to use vcron

- Use `vjob` when the task is splitting one finite workload into batches rather than recurring wall-clock scheduling.
- Use a durable external scheduler when jobs must survive process restarts, coordinate across replicas, or provide exactly-once semantics.
- Use `time.Ticker` or a simple loop for a single in-process interval task that does not need cron expression parsing.

## Parse and match cron expressions

```go
package main

import (
	"fmt"
	"time"

	"github.com/imajinyun/knifer-go/vcron"
)

func main() {
	p, err := vcron.NewPattern("* * * * *")
	if err != nil {
		panic(err)
	}

	now := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	fmt.Println(p.Match(now, false))
}
```

## Schedule functions with a local scheduler

```go
package main

import (
	"fmt"

	"github.com/imajinyun/knifer-go/vcron"
)

func main() {
	s := vcron.NewScheduler()
	id, err := s.ScheduleFunc("* * * * *", func() {
		fmt.Println("tick")
	})
	if err != nil {
		panic(err)
	}

	fmt.Println("scheduled", id)
	fmt.Println("removed", s.Remove(id))
}
```

## Customize scheduler options

```go
package main

import (
	"fmt"
	"time"

	"github.com/imajinyun/knifer-go/vcron"
)

func main() {
	loc := time.FixedZone("cst", 8*60*60)
	now := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	s := vcron.NewSchedulerWithOptions(
		vcron.WithLocation(loc),
		vcron.WithMatchSecond(true),
		vcron.WithIDGenerator(func() string { return "job-1" }),
		vcron.WithClock(func() time.Time { return now }),
		vcron.WithExecutor(func(fn func()) { fn() }),
	)

	id, err := s.ScheduleFunc("* * * * * *", func() {})
	if err != nil {
		panic(err)
	}
	fmt.Println(id, s.IsMatchSecond(), s.Config().Location == loc)
}
```

## Use the default scheduler facade

```go
package main

import (
	"fmt"

	"github.com/imajinyun/knifer-go/vcron"
)

func main() {
	id, err := vcron.CronScheduleFunc("* * * * *", func() {})
	if err != nil {
		panic(err)
	}

	fmt.Println(vcron.CronRemove(id))
}
```

## Related packages

- Use `vjob` for finite slice, map-key, or range batch workloads that are not wall-clock schedules.
- Use `vlog` when scheduled jobs need structured execution diagnostics.
- Use `verr` when job failures need collection, wrapping, or panic recovery.

## Benchmarks and trade-offs

Cron behavior is time- and concurrency-dependent, so validate scheduler behavior with focused tests:

```bash
go test ./internal/cron ./vcron
```

The main trade-offs are lifecycle and execution policy. Package-level schedulers are convenient but global; local schedulers are easier to isolate. Synchronous executors are deterministic for tests; asynchronous or pooled executors are better for production but require shutdown and capacity planning.

## FAQ

### Should libraries use the default scheduler?

Prefer returning or accepting a local `Scheduler`. The package-level scheduler is application-global state and can surprise tests or embedding applications.

### Does shutdown cancel running tasks?

Shutdown waits for launchers and running tasks or for the context to be canceled. Task functions need their own cancellation checks if they must stop early.

### Can tasks overlap?

Yes, if the executor starts a new run before the previous one finishes. Use task-level locks, leases, or idempotency when overlap is unsafe.

### When should I use second-level matching?

Use `WithMatchSecond(true)` only when the expression and operational need require seconds. Minute-level schedules are simpler and reduce accidental high-frequency execution.
