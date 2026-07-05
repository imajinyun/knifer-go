# vcli Quickstart

`vcli` provides lightweight command-line helpers for command execution, flag parsing, subcommand routing, and deterministic help text. It is dependency-free and intended for small tools or library code that needs CLI behavior without adopting a full framework.

## When to use vcli

| Scenario | Use `vcli` when | Prefer another tool when |
| --- | --- | --- |
| Run an external command | You need `context.Context`, separated args, captured output, optional timeout, and injectable runners for tests. | You need shell features such as pipes, glob expansion, or command substitution. Build those explicitly rather than passing user input to a shell. |
| Parse a few flags | A small utility needs deterministic `flag`-style parsing without a framework dependency. | You need completion generation, persistent flags, config-file binding, or large command trees; use Cobra/Viper instead. |
| Route subcommands | A library example or small binary needs predictable subcommand dispatch and captured stdout/stderr. | The command surface is user-facing and large enough to require rich help, aliases, completion, or plugin support. |
| Render help or color text | Tests need stable help text and explicit color policy. | Terminal styling is a core feature; use a dedicated color/table library. |

## Golden path APIs

The first-choice API set for this facade is kept in sync with `ai-context.json` and the generated tools catalog.

- `Output`
- `Colorize`
- `WithColorMode`
- `NewFlagParser`
- `RenderHelp`

## Which helper should I use?

Start with the helper that matches the boundary you are crossing: process execution, flag parsing, command routing, or presentation.

| Need | Use | Notes |
| --- | --- | --- |
| Execute a command and inspect stdout, stderr, exit code, or duration | `Run` | Pass args as `[]string`; `vcli` does not invoke a shell. Use `WithRunner` in tests. |
| Execute a command and only need stdout | `Output` | It is a thin wrapper over `Run`; errors still include command execution failures. |
| Avoid real process execution in tests | `WithRunner` with `RunnerFunc` | Assert the received `ExecRequest` instead of depending on host binaries. |
| Set command working directory, environment, stdin, timeout, or output cap | `WithDir`, `WithEnv`, `WithStdin`, `WithTimeout`, `WithMaxOutputBytes` | Keep these options visible at the call site so resource and environment boundaries are reviewable. |
| Parse command flags | `NewFlagParser` plus `String`, `Int`, `Bool`, `Duration` | Use `WithFlagOutput` to capture parse diagnostics and usage text. |
| Route a small subcommand tree | `Command.Execute` and `Command.Add` | Pass `WithStdout` and `WithStderr` to keep command output testable. |
| Render deterministic help or disable ANSI escapes | `RenderHelp`, `Colorize`, `WithColorMode` | Prefer `ColorNever` in generated examples and snapshot tests. |

## CLI safety checklist

- Pass command arguments as separate slice elements. Do not concatenate untrusted input into `sh -c` or another shell command.
- Use `WithRunner` for examples and tests so they do not depend on local binaries, PATH, locale, or operating-system behavior.
- Use `WithTimeout` for commands that may hang, especially wrappers around network, package-manager, or user-provided tools.
- Use `WithMaxOutputBytes` when stdout or stderr can be large or attacker-controlled.
- Treat `ExecResult.Stderr` as diagnostic output. Keep machine-readable output on stdout and errors/logs on stderr.
- Capture flag and command output with `WithFlagOutput`, `WithStdout`, and `WithStderr` instead of writing directly to process-global streams.

## When not to use vcli

- Use Cobra, Viper, or another CLI framework when the command surface needs completions, persistent flags, config-file binding, aliases, plugin loading, or generated command documentation.
- Use `os/exec` directly when you need low-level process control, custom pipes, streaming I/O, process groups, or platform-specific attributes.
- Avoid shell execution for untrusted input. If shell features are required, build and review the shell boundary explicitly instead of hiding it behind a helper.
- Use dedicated terminal UI, table, or color libraries when rich interactive output is a primary feature.
- Avoid external command execution in hot paths or request handlers unless timeout, output bounds, cancellation, and failure policy are explicit.

## Related packages

- Use `vconf` when CLI flags need to merge with files, environment variables, or remote configuration.
- Use `vlog` when command diagnostics need structured logging rather than plain stderr text.
- Use `verr` when command failures need wrapped errors, panic recovery, or aggregation.

## Benchmarks and trade-offs

Use focused benchmarks to compare command routing, flag parsing, output capture, and injected-runner overhead:

```bash
go test -bench=. -benchmem -run=^$ ./internal/cli ./vcli
```

`vcli` keeps small tools dependency-free and easy to test, but it intentionally does not provide shell semantics, completions, config binding, or a large CLI framework. For hot command-routing paths, measure help rendering and flag parsing separately from external process execution; starting a process will dominate helper overhead.

Output capture and `WithMaxOutputBytes` add buffering and bounds checks. Keep them enabled when command output can be large or untrusted, and inject `RunnerFunc` in tests to avoid measuring host binaries or PATH behavior.

## FAQ

### Does vcli replace Cobra?

No. `vcli` is intentionally small. Use it for small utilities, library examples, and deterministic tests. Use Cobra/Viper for large user-facing CLIs that need completions, persistent flags, config layering, aliases, or generated command docs.

### Does Run invoke a shell?

No. `Run` passes the command name and argument slice to the runner. The default runner uses `exec.CommandContext`, so shell metacharacters are not expanded unless you explicitly choose to execute a shell.

### How should I test code that uses vcli?

Inject a `RunnerFunc` with `WithRunner`, and capture command I/O with in-memory buffers. This keeps tests hermetic and avoids depending on installed tools or platform-specific command output.

## Run a command with an injected runner

```go
package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/imajinyun/knifer-go/vcli"
)

func main() {
	runner := vcli.RunnerFunc(func(ctx context.Context, req vcli.ExecRequest) (vcli.ExecResult, error) {
		return vcli.ExecResult{Stdout: req.Name + " " + strings.Join(req.Args, " ")}, nil
	})

	result, err := vcli.Run(context.Background(), "echo", []string{"hello"}, vcli.WithRunner(runner))
	if err != nil {
		panic(err)
	}
	fmt.Println(result.Stdout)
}
```

`vcli.Run` accepts a command name plus an argument slice. It does not invoke a shell by default. Use `WithRunner` in tests to avoid starting real processes.

## Parse flags

```go
package main

import (
	"fmt"

	"github.com/imajinyun/knifer-go/vcli"
)

func main() {
	parser := vcli.NewFlagParser("serve")
	port := parser.Int("port", 8080, "port to bind")
	debug := parser.Bool("debug", false, "enable debug")
	result, err := parser.Parse([]string{"--port", "9090", "--debug", "api"})
	if err != nil {
		panic(err)
	}
	fmt.Println(*port, *debug, result.Args)
}
```

## Route subcommands

```go
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/imajinyun/knifer-go/vcli"
)

func main() {
	root := &vcli.Command{Name: "app", Usage: "app <command>", Summary: "demo app"}
	root.Add(&vcli.Command{Name: "hello", Summary: "print greeting", Run: func(ctx context.Context, inv *vcli.Invocation) error {
		_, _ = fmt.Fprintf(inv.Stdout, "hello %s\n", inv.Args[0])
		return nil
	}})

	if err := root.Execute(context.Background(), []string{"hello", "gopher"}, vcli.WithStdout(os.Stdout), vcli.WithStderr(os.Stderr)); err != nil {
		panic(err)
	}
}
```

## Render help without color

```go
package main

import (
	"fmt"

	"github.com/imajinyun/knifer-go/vcli"
)

func main() {
	root := &vcli.Command{Name: "app", Usage: "app <command>", Summary: "demo app"}
	root.Add(&vcli.Command{Name: "serve", Summary: "start server"})
	fmt.Print(vcli.RenderHelp(root, vcli.WithColorMode(vcli.ColorNever)))
}
```
