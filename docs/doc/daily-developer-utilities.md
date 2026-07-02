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

## Cookbook

| Task | Facades | Short path | Boundary |
| --- | --- | --- | --- |
| env-driven command execution | `vcli`, `vsys`, `vconf` | read environment with `vsys`, parse config with `vconf`, execute through `vcli.Output` with an injected runner in tests | Keep command arguments as slices and avoid shell concatenation for untrusted input. |
| config-backed file workflow | `vconf`, `vfile`, `vlog` | parse local config, read or write small files, log the selected path and explicit error | Keep path policy visible before calling file helpers. |
| network diagnostics report | `vnet`, `vsys`, `vlog` | inspect host/system data, check IP/CIDR/port details, emit a structured console log | Treat DNS, dial, and interface data as environment-specific evidence. |
| CLI support bundle | `vcli`, `vsys`, `vfile`, `vlog` | render command help, dump system info, collect bounded files, write diagnostics | Keep generated evidence out of source control unless it is stable documentation. |
| local batch job runner | `vjob`, `vcli`, `vlog` | split work into typed jobs, call injected commands, merge results, log failures | Use context-aware APIs when jobs may block or be canceled. |
| filesystem cleanup preview | `vfile`, `vsys`, `vlog` | inspect candidate paths, log planned actions, execute only after caller-owned policy checks | `vfile` helpers do not authorize deletion by themselves. |
| lightweight service smoke script | `vconf`, `vnet`, `vhttp`, `vlog` | read endpoint config, validate host/port, call safe HTTP helpers, log status | Use Safe/E/WithOptions flows when endpoint config crosses a trust boundary. |

## Decision Rules

- Use the Go standard library when a direct call is shorter and there is no
  shared project policy to preserve.
- Use `gookit/goutil` when a project already accepts its broad helper-package
  style and does not need `knifer-go` safety or governance boundaries.
- Use `knifer-go` when daily utilities should sit beside safe HTTP, URL,
  crypto, JWT, file, config, database, and provider-injected helpers.
- Keep untrusted filesystem paths, command arguments, network addresses, and
  secrets in `Safe`, `E`, context-aware, or `WithOptions` flows when available.

Machine-readable boundaries:

- env-driven command execution
- config-backed file workflow
- network diagnostics report
- CLI support bundle
- local batch job runner
- filesystem cleanup preview
- lightweight service smoke script
- vtest and vdump are planned lanes
- daily utilities should stay beside safety-focused facades
- no resident background utility process
- Safe/E/WithOptions flows for trust boundaries

## Planned Lane

vtest is a planned lane, not a current public facade. Until it lands, use
standard Go tests, package-local helpers, and provider injection in the target
facade. Do not document `vtest` as available API until it appears in the public
facade catalog.

`vdump` is also a planned lane, not a current public facade. Until it lands, use
`vsys.DumpSystemInfo`, targeted logs, and caller-owned diagnostic writers.
Daily utilities should stay beside safety-focused facades without becoming a
resident background utility process.
