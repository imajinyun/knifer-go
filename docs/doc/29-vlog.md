# vlog Quickstart

`vlog` provides console logging facades with package-level static logs, named logger lookup, colored output, log levels, and custom output targets.

## Which helper should I use?

| Goal | Start with | Notes |
| --- | --- | --- |
| Emit a quick package-level message | `Info`, `Warn`, `ErrorLog`, `Debug`, `Trace` | Uses the static default logger and package-level level threshold. Keep this for applications and small tools. |
| Format a message | `Infof`, `Warnf`, `Errorf`, `LogAt` | Prefer formatted helpers when the message shape is stable and arguments are cheap to compute. |
| Include an error with a level | `LogAtE` | Keeps the error value visible to the logger instead of flattening it into the format string. |
| Use a named logger | `Logger` or `DefaultLogger` | Named loggers are cached by default, so repeated lookups avoid rebuilding console state. |
| Customize one lookup or static call | `LoggerWithOptions`, `InfoWithOptions`, `LogAtWithOptions` | Pass `WithLoggerCache(false)` for isolated tests or request-scoped output. |
| Build a deterministic console logger | `NewConsoleLogWithOptions` | Inject `WithLogOutput`, `WithLogClock`, and `WithLogTimeLayout` for reproducible examples and tests. |
| Add ANSI colors | `NewConsoleColorLogWithOptions` | Use `WithLogColorFactory` when tests need stable escape sequences or production wants a custom palette. |
| Replace the package-level logger implementation | `SetLogFactory` | Application-level integration point; libraries should prefer per-call options to avoid global side effects. |
| Change console verbosity globally | `SetLogLevel` / `GetLogLevel` | Set once during application startup, not inside library code or hot paths. |

## Logging safety checklist

- Prefer `LoggerWithOptions` or `NewIsolatedLogger` in tests so captured output, clocks, and cache state do not leak between cases.
- Avoid mutating package-level state (`SetLogFactory`, `SetLogLevel`, `SetLogColorFactory`) from libraries; let the binary configure logging at the boundary.
- Route normal output and error output separately with `WithLogOutput` when command-line tools need stdout to remain machine-readable.
- Inject `WithLogClock` for golden tests and executable examples; wall-clock timestamps make output flaky.
- Do not log secrets, tokens, credentials, or full request bodies. Redact before calling the facade.
- Check the current level before doing expensive message construction outside the logger; the facade can suppress output but cannot undo argument work already done.

## Use static logging functions

```go
package main

import "github.com/imajinyun/go-knifer/vlog"

func main() {
	vlog.SetLogLevel(vlog.LogLevelDebug)
	vlog.Debug("debug message")
	vlog.Infof("hello %s", "go-knifer")
	vlog.Warn("watch this")
}
```

## Create a console logger and customize output

```go
package main

import (
	"bytes"
	"fmt"
	"time"

	"github.com/imajinyun/go-knifer/vlog"
)

func main() {
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	fixed := time.Date(2024, 4, 5, 6, 7, 8, 0, time.UTC)

	log := vlog.NewConsoleLogWithOptions("demo",
		vlog.WithLogOutput(out, errOut),
		vlog.WithLogClock(func() time.Time { return fixed }),
		vlog.WithLogTimeLayout(time.RFC3339),
		vlog.WithLogLevel(vlog.LogLevelInfo),
	)
	log.Info("ready")
	log.Error("failed")

	fmt.Println(out.String())
	fmt.Println(errOut.String())
}
```

## Use a colored logger

```go
package main

import (
	"bytes"
	"fmt"

	"github.com/imajinyun/go-knifer/vlog"
)

func main() {
	out := &bytes.Buffer{}
	log := vlog.NewConsoleColorLogWithOptions("color",
		vlog.WithLogOutput(out, &bytes.Buffer{}),
		vlog.WithLogColorFactory(func(level vlog.Level) string { return "\033[36m" }),
	)
	log.Info("colored")
	fmt.Println(out.String())
}
```

## When not to use vlog

- Use `log/slog`, zap, zerolog, or a platform logger when you need structured fields, sampling, trace correlation, or JSON log contracts.
- Use `verr` when the main task is panic recovery, stack capture, error aggregation, or Sentry/logrus setup; `vlog` is the lightweight console facade.
- Avoid package-level static helpers in reusable libraries that must not change application logging policy.
- Avoid colored console output when logs are parsed by machines unless the output sink explicitly supports ANSI escapes.

## Related packages

- Use `verr` when logs should include wrapped errors, stack capture, panic recovery, or collected failures.
- Use `vcli` when command-line tools need deterministic stdout/stderr capture alongside logging.
- Use `vjson` when structured log payloads or fixtures need JSON formatting and inspection.

## Benchmarks and trade-offs

- Cached `Logger` lookups trade a small amount of package-level state for lower allocation and construction cost on repeated named logging.
- `WithLoggerCache(false)` and `NewIsolatedLogger` are easier to reason about in tests, but they rebuild logger state for each call.
- Colored output adds formatting work and bytes on the wire; keep it for terminals, not high-volume machine logs.
- `WithLogClock` and fixed time layouts make tests deterministic without changing production behavior.
- Package-level level checks suppress emitted lines, but callers should still avoid building large strings or serializing objects unless the log will be used.

## FAQ

### Should a library call `SetLogLevel` or `SetLogFactory`?

No. Treat those as application startup configuration. Libraries should accept a logger or use `LoggerWithOptions`/`NewIsolatedLogger` so callers keep control of global logging behavior.

### How do I make log output deterministic in tests?

Create a logger with `NewConsoleLogWithOptions`, pass `WithLogOutput` buffers, `WithLogClock` with a fixed time, and a stable `WithLogTimeLayout`. Disable cache with `WithLoggerCache(false)` when using lookup helpers.

### Why are error logs written to a different writer?

Console loggers can route normal and error streams separately. This lets CLI tools keep stdout for data while warnings and errors go to stderr.

### When should I use `LogAtE` instead of `Errorf`?

Use `LogAtE` when you have an `error` value and want the logger to receive it explicitly. Use `Errorf` for pure formatted text.

## Choose logger configuration for a single call

```go
package main

import (
	"bytes"
	"fmt"

	"github.com/imajinyun/go-knifer/vlog"
)

func main() {
	out := &bytes.Buffer{}
	log := vlog.LoggerWithOptions("request",
		vlog.WithLoggerCache(false),
		vlog.WithLoggerConsoleOptions(vlog.WithLogOutput(out, &bytes.Buffer{})),
	)
	log.Info("request started")

	vlog.InfoWithOptions([]vlog.LoggerOption{
		vlog.WithLoggerCache(false),
		vlog.WithLoggerConsoleOptions(vlog.WithLogOutput(out, &bytes.Buffer{})),
	}, "static call")

	fmt.Println(out.String())
}
```
