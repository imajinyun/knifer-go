# verr Quickstart

`verr` provides error aggregation, panic recovery, stack capture, and logrus/Sentry initialization helpers for centralized handling of errors from synchronous or asynchronous tasks.

## Which helper should I use?

| Goal | Start with | Notes |
| --- | --- | --- |
| Convert panics into errors around one function | `Recover` / `RecoverWithoutError` | Captures a panic as `PanicError` and logs with the default error-level logger. |
| Wrap a fallible function with fluent logging | `Wrap(...).WithWarnf(...).Exec(ctx)` | Use when the call site wants to choose the log level and message before execution. |
| Aggregate errors from concurrent tasks | `NewCollector` / `NewCollectorWithOptions` | `Collector` recovers panics, logs failures, and returns an aggregated error. |
| Launch collector work with a custom scheduler | `WithCollectorRunner` | Useful for deterministic tests that run async work synchronously. |
| Wait with an injectable timer | `WithCollectorTimerFactory` / `WithTimerFactory` | Keeps timeout tests hermetic without sleeping on wall-clock time. |
| Check an aggregate for a sentinel error | `ErrorIs` | Extends `errors.Is` behavior across multi-error members. |
| Capture or format stack frames | `GetStackTraceWithOptions`, `GetStackWithOptions` | Use `WithStackDepth`, `WithStackSkip`, and provider hooks for deterministic frame capture. |
| Initialize logrus/Sentry globally | `InitWithOptions` | Application startup helper; inject factories in tests to avoid real Sentry side effects. |
| Build an isolated logrus logger | `NewIsolatedLogrusWithOptions` | Prefer for tests and libraries that must not mutate global logrus state. |
| Fail fast on an error | `MustExitWithOptions` | Defaults to panic-on-error; inject `WithExitPanicFunc` and `WithExitLogFunc` in tests. |

## Error-handling safety checklist

- Recover panics at goroutine and task boundaries, not deep inside pure helper functions where panics should surface during tests.
- Use `ErrorIs` for collector results or joined errors so sentinel checks do not miss nested members.
- Inject `WithCollectorRunner` and timer factories in tests; real goroutines and real timers make order and duration assertions flaky.
- Prefer `NewIsolatedLogrusWithOptions` in tests or libraries. `InitWithOptions` configures global logrus/Sentry integration.
- Bound stack capture with `WithStackDepth` when stack traces are emitted frequently or stored for later inspection.
- Do not send real Sentry events from unit tests. Use `WithSentryClient`, `WithSentryClientFactory`, `WithSentryHookFactory`, and `WithLogHookAdder` to replace external effects.
- Avoid logging sensitive values in panic messages or recovery formats; recovery helpers preserve enough context to leak secrets if callers include them.

## Recover panics and return errors

```go
package main

import (
	"fmt"

	"github.com/imajinyun/go-knifer/verr"
)

func main() {
	err := verr.Recover(func() error {
		panic("boom")
	}, "running job")

	fmt.Printf("%T\n", err)
	fmt.Println(verr.GetStack(err) != "")
}
```

## Wrap fallible functions with Wrapper

```go
package main

import (
	"context"
	"errors"
	"fmt"

	"github.com/imajinyun/go-knifer/verr"
)

func main() {
	want := errors.New("write failed")
	err := verr.Wrap(func() error { return want }).WithWarnf("save user").Exec(context.Background())

	fmt.Println(verr.ErrorIs(err, want))
}
```

## Aggregate errors from multiple asynchronous tasks

```go
package main

import (
	"errors"
	"fmt"

	"github.com/imajinyun/go-knifer/verr"
)

func main() {
	c := verr.NewCollector()
	want := errors.New("task failed")

	c.GoRun(func() error { return nil }, "task one")
	c.GoRun(func() error { return want }, "task two")

	err := c.Error()
	fmt.Println(verr.ErrorIs(err, want))
}
```

## Capture stack frames

```go
package main

import (
	"fmt"

	"github.com/imajinyun/go-knifer/verr"
)

func main() {
	stack := verr.GetStackTraceWithOptions(
		verr.WithStackSkip(0),
		verr.WithStackDepth(4),
	)
	fmt.Println(len(stack) > 0)
}
```

## Initialize Sentry without global side effects in tests

`InitWithOptions` accepts Sentry factories so production code can use
`sentry-go` while tests inject isolated clients and hook registration. Prefer
`WithSentryClientOptions`, `WithSentryClient`, or `WithSentryClientFactory`.

```go
package main

import (
	"fmt"

	"github.com/getsentry/sentry-go"
	"github.com/imajinyun/go-knifer/verr"
	"github.com/sirupsen/logrus"
)

type memoryHook struct{}

func (memoryHook) Levels() []logrus.Level { return []logrus.Level{logrus.ErrorLevel} }

func (memoryHook) Fire(*logrus.Entry) error { return nil }

func main() {
	var registered bool

	verr.InitWithOptions(
		verr.WithSentryDSN("https://public@example.invalid/1"),
		verr.WithSentryClientOptions(sentry.ClientOptions{Environment: "test"}),
		verr.WithSentryClientFactory(func(options sentry.ClientOptions) (*sentry.Client, error) {
			return sentry.NewClient(options)
		}),
		verr.WithSentryHookFactory(func(*sentry.Client, []logrus.Level) (logrus.Hook, error) {
			return memoryHook{}, nil
		}),
		verr.WithLogHookAdder(func(logrus.Hook) { registered = true }),
	)

	fmt.Println(registered)
}
```

## When not to use verr

- Use ordinary `error` returns and `%w` wrapping for simple synchronous flows that do not need recovery, aggregation, or stack capture.
- Use `context` cancellation, `errgroup`, or worker-pool control when the primary concern is lifecycle management rather than error collection.
- Use the application's observability stack directly when logs, traces, and metrics already have a structured error contract.
- Avoid `MustExit` in libraries; return errors so callers decide whether to panic, retry, or continue.

## Related packages

- Use `vlog` when errors should be emitted through named loggers, custom outputs, or structured log flows.
- Use `vjson` when error payloads need JSON formatting for tests, APIs, or diagnostics.
- Use `vhttp` or `vresty` when wrapped errors originate from HTTP client boundaries.

## Benchmarks and trade-offs

- Stack capture is useful for diagnostics but allocates and walks runtime metadata. Capture stacks at error boundaries, not on every successful operation.
- `Collector.GoRun` simplifies panic-safe fan-out but introduces synchronization and scheduling overhead compared with a direct function call.
- Custom runner and timer providers are slightly more setup, but they make tests deterministic and faster than sleeping.
- Global `InitWithOptions` is convenient at application startup, while isolated loggers avoid hidden state and cross-test interference.
- Frame metadata caching can reduce repeated stack formatting cost; reset or disable it when tests need to observe provider calls exactly.

## FAQ

### Does `Recover` hide programmer bugs?

It can if used too broadly. Put recovery at process, request, goroutine, or job boundaries so the service can keep running while tests and inner functions still expose unexpected panics.

### How do I test collector concurrency deterministically?

Construct the collector with `WithCollectorRunner(func(f func()) { f() })` to run submitted work inline, and use `WithCollectorTimerFactory` when testing wait behavior.

### Why should I use `ErrorIs` instead of `errors.Is`?

`ErrorIs` delegates normal checks but also walks aggregated errors produced by the collector, so sentinel matches are not lost after multiple task failures are combined.

### Should `InitWithOptions` be called by package initialization code?

No. Call it from application startup. Package initialization should avoid mutating global loggers or registering external reporting hooks.
