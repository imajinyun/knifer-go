# vcli Quickstart

`vcli` provides lightweight command-line helpers for command execution, flag parsing, subcommand routing, and deterministic help text. It is dependency-free and intended for small tools or library code that needs CLI behavior without adopting a full framework.

## Run a command with an injected runner

```go
package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/imajinyun/go-knifer/vcli"
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

	"github.com/imajinyun/go-knifer/vcli"
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

	"github.com/imajinyun/go-knifer/vcli"
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

	"github.com/imajinyun/go-knifer/vcli"
)

func main() {
	root := &vcli.Command{Name: "app", Usage: "app <command>", Summary: "demo app"}
	root.Add(&vcli.Command{Name: "serve", Summary: "start server"})
	fmt.Print(vcli.RenderHelp(root, vcli.WithColorMode(vcli.ColorNever)))
}
```

