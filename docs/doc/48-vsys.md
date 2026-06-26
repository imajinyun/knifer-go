# vsys Quickstart

`vsys` provides system information, Go runtime information, process metrics, environment reads, and system-info dump helpers, with option-based data provider injection.

## Which helper should I use?

| Goal | Start with | Notes |
| --- | --- | --- |
| Read cached host information | `SystemHostInfo` | Uses package-level cache for stable host metadata. Call `ResetInfoCache` when tests need a fresh singleton. |
| Read uncached or injected host information | `SystemHostInfoWithOptions` / `SysHostInfoWithOptions` | Use `WithHostNameFunc`, `WithHostInterfaceAddrsFunc`, or `WithHostAddressFunc` for deterministic tests. |
| Read OS metadata | `SystemOsInfo` / `SystemOsInfoWithOptions` | Inject OS name, arch, version, and separator providers when the host platform should not affect assertions. |
| Read current user metadata | `SystemUserInfo` / `SystemUserInfoWithOptions` | Use `WithCurrentUserFunc`, `WithUserEnvLookup`, `WithWorkingDirFunc`, and `WithTempDirFunc` to avoid machine-specific paths. |
| Read Go runtime metadata | `SystemGoInfo` / `SystemGoInfoWithOptions` | Inject `WithGoVersionFunc`, `WithGoRootFunc`, or `WithGoEnvOutputFunc` when `go env` should not be executed. |
| Read runtime memory snapshots | `SystemRuntimeInfoWithOptions`, `TotalMemoryWithOptions`, `FreeMemoryWithOptions` | `WithReadMemStatsFunc` lets tests provide fixed memory counters. |
| Read process scalar values | `CurrentPIDWithOptions`, `TotalGoroutineCountWithOptions` | Inject `WithPIDFunc` or `WithProcessNumGoroutineFunc` for reproducible examples. |
| Read environment values with defaults | `EnvOrDefaultWithOptions`, `EnvIntWithOptions`, `EnvBoolWithOptions` | Use `WithEnvLookupFunc` and parser hooks to avoid changing process environment in tests. |
| Dump all system information | `DumpSystemInfoWithOptions` | Write to an explicit `io.Writer` and pass `WithDump*Options` for deterministic output. |

## System information safety checklist

- Prefer `WithOptions` variants in tests; hostnames, users, paths, memory counters, and goroutine counts vary by machine.
- Call `ResetInfoCache` only in tests or controlled setup. Cached singleton data should not be reset concurrently with production readers.
- Use `EnvOrDefault`, `EnvInt`, and `EnvBool` when absence or invalid input should fall back safely; reserve `Env` for required string values.
- Redirect missing-environment warnings with `WithEnvWarningWriter` if stderr output would break tests or CLI protocols.
- Inject `WithGoEnvOutputFunc` or `WithGoRootFunc` when running in sandboxes that may not have the `go` binary available.
- Avoid dumping raw environment-derived or user path information into public logs unless the output is reviewed for sensitive data.

## Read host, OS, and Go information

```go
package main

import (
	"fmt"
	"runtime"

	"github.com/imajinyun/knifer-go/vsys"
)

func main() {
	host := vsys.SystemHostInfoWithOptions(vsys.WithHostNameFunc(func() (string, error) {
		return "dev-host", nil
	}))
	osInfo := vsys.SystemOsInfo()
	goInfo := vsys.SystemGoInfoWithOptions(vsys.WithGoVersionFunc(runtime.Version))

	fmt.Println(host.Name)
	fmt.Println(osInfo.Name)
	fmt.Println(goInfo.Version != "")
}
```

## Read process and runtime metrics

```go
package main

import (
	"fmt"

	"github.com/imajinyun/knifer-go/vsys"
)

func main() {
	fmt.Println(vsys.CurrentPIDWithOptions(vsys.WithPIDFunc(func() int { return 1234 })))
	fmt.Println(vsys.TotalGoroutineCountWithOptions(vsys.WithProcessNumGoroutineFunc(func() int { return 8 })))
	fmt.Println(vsys.TotalMemory() >= 0)
	fmt.Println(vsys.FreeMemory() >= 0)
}
```

## Read environment variables

```go
package main

import (
	"fmt"

	"github.com/imajinyun/knifer-go/vsys"
)

func main() {
	lookup := func(key string) (string, bool) {
		values := map[string]string{"PORT": "8080", "DEBUG": "true"}
		v, ok := values[key]
		return v, ok
	}

	fmt.Println(vsys.EnvOrDefaultWithOptions("APP", "knifer-go", vsys.WithEnvLookupFunc(lookup)))
	fmt.Println(vsys.EnvIntWithOptions("PORT", 80, vsys.WithEnvLookupFunc(lookup)))
	fmt.Println(vsys.EnvBoolWithOptions("DEBUG", false, vsys.WithEnvLookupFunc(lookup)))
}
```

## Dump system information to a writer

```go
package main

import (
	"bytes"
	"fmt"

	"github.com/imajinyun/knifer-go/vsys"
)

func main() {
	var out bytes.Buffer
	vsys.DumpSystemInfoWithOptions(&out,
		vsys.WithDumpHostOptions(vsys.WithHostNameFunc(func() (string, error) { return "docs", nil })),
	)

	fmt.Println(out.Len() > 0)
}
```

## When not to use vsys

- Use the standard library directly (`os`, `runtime`, `os/user`) when you need one scalar value and do not need facade consistency or provider injection.
- Use a metrics or observability library for long-lived runtime monitoring; `vsys` snapshots are not a metrics registry.
- Avoid `DumpSystemInfo` in user-facing responses or public bug reports unless sensitive fields are reviewed.
- Do not use cached singleton helpers when tests require fresh values after changing provider state; use `WithOptions` helpers instead.

## Related packages

- Use `vcli` when system information must be gathered from external commands with timeouts and captured output.
- Use `vconf` when system-derived defaults should be merged with files, environment variables, or remote config.
- Use `vlog` when host, runtime, or process details are emitted as diagnostics.

## Benchmarks and trade-offs

- Cached host, OS, user, and Go helpers reduce repeated system calls but can return stale values after environment or working-directory changes.
- Runtime memory helpers call `runtime.ReadMemStats`, which can be more expensive than returning cached metadata. Sample deliberately instead of on every hot-path request.
- Environment helpers are cheap, but warning output and parsing hooks are observable side effects; inject them in tests.
- `DumpSystemInfoWithOptions` centralizes collection, but it gathers multiple categories and writes formatted output, so keep it out of latency-sensitive paths.
- Provider injection adds option setup at call sites but removes dependence on the developer's machine, CI host, and shell environment.

## FAQ

### Why are some helpers cached and others refreshed?

Host, OS, user, and Go metadata usually change rarely, so cached helpers are convenient. Runtime memory and goroutine counts are point-in-time process metrics, so they are refreshed.

### How do I make environment-variable tests hermetic?

Use `WithEnvLookupFunc` instead of mutating the process environment. Pair it with `WithEnvIntParser`, `WithEnvBoolParser`, or `WithEnvWarningWriter` when parsing or warnings are part of the assertion.

### When should I call `ResetInfoCache`?

Use it in tests after exercising cached singleton helpers. Avoid calling it in production request paths because other goroutines may be reading the same cached values.

### How can I avoid invoking `go env`?

Pass `WithGoRootFunc` when you already know the GOROOT value, or `WithGoEnvOutputFunc` when you want to replace the command runner used by the default collector.
