# vconf Quickstart

`vconf` reads, parses, and manages grouped configuration, with support for setting/properties, YAML, TOML, profile overrides, environment expansion, schema validation, and file watching.

## Which helper should I use?

Choose the helper by source format, layering needs, and how strictly configuration must be validated before use.

| Need | Use | Notes |
| --- | --- | --- |
| Parse inline setting/properties text | `Parse`, `ParseBytes` | Good for tests, generated config, and simple key/value files. |
| Parse by file extension | `ParseByExt`, `ParseByExtWithOptions` | Keeps format dispatch explicit when callers accept multiple config formats. |
| Parse YAML or TOML | `ParseYAML`, `ParseYAMLFull`, `ParseTOML` | Use the full parser variants when nested structures matter. |
| Load one or more files | `Load`, `LoadFiles`, `LoadWithOptions` | Use ordered file lists to make precedence reviewable. |
| Apply environment or profile overrides | `GetExpandedWithOptions`, `ApplyProfile`, `LoadProfile` | Inject env lookup in tests for deterministic expansion. |
| Bind configuration to structs | `Bind`, `BindGroup`, bind options, `WithBindDecodeHook` | Prefer binding before application startup code uses configuration values; keep custom conversions local to the bind call. |
| Validate configuration | `SchemaFromStruct`, validation helpers | Validate required fields, ranges, and type expectations before starting long-running work. |
| Watch a file for changes | `Watch`, `WatchWithOptions` | Ensure callbacks are idempotent and safe to run more than once. |
| Load remote configuration | `LoadRemoteSafe`, `LoadRemoteSafeWithOptions` | Prefer safe remote helpers for URLs from config, users, or service discovery. |

## Configuration safety checklist

- Keep configuration precedence explicit: defaults, files, profiles, environment expansion, and remote sources should be reviewable in order.
- Use safe remote loading for any URL that is not a compile-time constant owned by the application.
- Inject environment lookup in tests instead of depending on the host process environment.
- Validate required fields and ranges before starting services, opening network listeners, or launching background workers.
- Treat decrypted or expanded values as secrets when they contain credentials; do not log raw configuration maps.
- Make file watchers resilient: callbacks should tolerate partial writes, repeated events, and invalid intermediate config.

## Bind hook and key-path errors

`WithBindDecodeHook` adds a per-call conversion hook for binding text values into custom destination types such as `time.Time`, enum wrappers, or domain-specific identifiers. Hooks receive the source type, destination type, and current value. Return the original string unchanged when the hook does not handle the conversion. Avoid global bind registries; configuration conversion policy should be visible beside the `BindWithOptions` or `BindGroupWithOptions` call.

Bind errors include the configuration key path. Nested struct binding reports keys such as `server.port`, and schema validation reports the failing key from the rule. Tests for new bind paths should assert both the error code and the key path so users can fix configuration without stepping through reflection code.

## When not to use vconf

- Use Viper, Cobra integration, or a larger configuration framework when you need many providers, live precedence stacks, or CLI flag binding across a large application.
- Use typed constructor parameters instead of configuration maps for library APIs; libraries should not read process-wide config implicitly.
- Use a secrets manager instead of `Base64Decrypt` or environment expansion for production credentials that require rotation, audit, and access control.
- Avoid remote loading for untrusted URLs unless `LoadRemoteSafeWithOptions` and an explicit URL policy cover the boundary.
- Avoid file watching when reload callbacks cannot be made idempotent or when partial writes would leave the application in an unsafe state.

## Related packages

- Use `vfile` when configuration loading first requires safe path or filesystem handling.
- Use `vjson` when JSON configuration needs direct formatting, inspection, or path access.
- Use `vbean` and `vform` when decoded configuration must be bound and validated explicitly.
- Use `vobj` only for generic nil/empty/default checks after configuration has already been loaded and bound.

## Boundary with vbean and vobj

`vconf` owns configuration sources and policy: file formats, file lists, profile overlays, environment expansion, schema validation, remote loading, and watch callbacks. It should not grow generic object utilities or struct-copy behavior just because configuration is often bound into structs.

Use `vbean` when a loaded map or struct must be copied into another Go shape with tag matching, weak conversion, or matched/unused metadata. Use `vobj` for generic optional-value checks around the already-bound result. Keep the sequence reviewable: load and validate with `vconf`, bind/map with `vbean` if needed, then apply typed or `vobj` checks in application code.

## Benchmarks and trade-offs

Use local benchmarks to compare parsing, binding, schema validation, environment expansion, and safe remote-loading overhead:

```bash
go test -bench=. -benchmem -run=^$ ./internal/conf/... ./vconf
```

Direct typed constructors are the simplest and fastest option for small programs. `vconf` is useful when configuration needs grouping, profile overlays, schema validation, and repeatable parsing across multiple formats.

Safety and flexibility add work. Full YAML parsing, schema reflection, remote URL validation, and file watch loops are easier to review when centralized, but they should still be measured and scoped to startup or explicit reload paths.

## FAQ

### Does vconf replace Viper?

No. `vconf` is a lightweight grouped configuration helper for common parsing, binding, profile, and validation workflows. Use Viper when you need broad ecosystem features such as Cobra integration, remote providers, or complex precedence stacks.

### When should I use ParseYAMLFull instead of ParseYAML?

Use `ParseYAMLFull` when nested YAML structures and standard YAML behavior matter. Use `ParseYAML` for the smaller supported subset when simple grouped config is enough.

### How should I test environment expansion?

Use `WithEnvLookup` to inject deterministic values. This keeps tests independent of the developer machine, CI environment, and secret-bearing process variables.

## Parse TOML and read grouped values

```go
package main

import (
	"fmt"

	"github.com/imajinyun/knifer-go/vconf"
)

func main() {
	c, err := vconf.ParseTOML(`
name = "demo"
[server]
port = 8080
debug = true
`)
	if err != nil {
		panic(err)
	}

	fmt.Println(c.Get("name"))
	fmt.Println(c.GetIntByGroup("server", "port", 0))
	fmt.Println(c.GetBoolByGroup("server", "debug", false))
}
```

## Expand environment variables

```go
package main

import (
	"fmt"

	"github.com/imajinyun/knifer-go/vconf"
)

func main() {
	c, err := vconf.Parse("base=http://${ENV:HOST}\n")
	if err != nil {
		panic(err)
	}

	value := c.GetExpandedWithOptions("base", vconf.WithEnvLookup(func(name string) string {
		if name == "HOST" {
			return "localhost:8080"
		}
		return ""
	}))
	fmt.Println(value)
}
```

## Bind to a struct

```go
package main

import (
	"fmt"

	"github.com/imajinyun/knifer-go/vconf"
)

type Server struct {
	Port  int      `conf:"port"`
	Debug bool     `conf:"debug"`
	Tags  []string `conf:"tags"`
}

func main() {
	c, err := vconf.ParseTOML(`
[server]
port = 8080
debug = true
tags = ["api", "admin"]
`)
	if err != nil {
		panic(err)
	}

	var server Server
	if err := c.BindGroup("server", &server); err != nil {
		panic(err)
	}
	fmt.Printf("%+v\n", server)
}
```

## Apply profile overrides

```go
package main

import (
	"fmt"

	"github.com/imajinyun/knifer-go/vconf"
)

func main() {
	c, err := vconf.ParseTOML(`
[server]
port = 8080
[profile.prod.server]
port = 9090
`)
	if err != nil {
		panic(err)
	}

	prod := c.ApplyProfile("prod")
	fmt.Println(prod.GetByGroup("server", "port"))
}
```
