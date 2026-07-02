# Daily Developer Utilities

Use this page when a project wants the `gookit/goutil` style mental model:
small daily helpers for local development, scripts, diagnostics, and operational
glue. In `knifer-go`, those helpers stay in focused facade packages instead of
one broad mixed package.

## Entry Points

| Workflow | Start here | Boundary |
| --- | --- | --- |
| CLI commands and terminal output | `vcli` | Use for typed flag parsing, subcommand routing, deterministic help text, command execution, and color output. |
| System and runtime inspection | `vsys` | Use for host, OS, user, process, environment, Go runtime, and dump helpers. |
| File and IO tasks | `vfile` | Use for file reads, writes, copy, lines, explicit errors, provider-backed tests, and bounded IO. |
| Network diagnostics | `vnet` | Use for IP, CIDR, local port, interface, DNS, TLS, ping, dial, and multipart helpers. |
| Local job orchestration | `vjob` | Use for batch work, typed map jobs, context cancellation, and serialized result merging. |
| Logging while scripting | `vlog` | Use for console logging, levels, isolated loggers, color controls, and per-call options. |

## Decision Rules

- Use the Go standard library when a direct call is shorter and there is no
  shared project policy to preserve.
- Use `gookit/goutil` when a project already accepts its broad helper-package
  style and does not need `knifer-go` safety or governance boundaries.
- Use `knifer-go` when daily utilities should sit beside safe HTTP, URL,
  crypto, JWT, file, config, database, and provider-injected helpers.
- Keep untrusted filesystem paths, command arguments, network addresses, and
  secrets in `Safe`, `E`, context-aware, or `WithOptions` flows when available.

## Planned Lane

vtest is a planned lane, not a current public facade. Until it lands, use
standard Go tests, package-local helpers, and provider injection in the target
facade. Do not document `vtest` as available API until it appears in the public
facade catalog.
