package vcli_test

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/imajinyun/knifer-go/vcli"
)

func ExampleRun() {
	runner := vcli.RunnerFunc(func(ctx context.Context, req vcli.ExecRequest) (vcli.ExecResult, error) {
		return vcli.ExecResult{Stdout: req.Name + " " + strings.Join(req.Args, " ")}, nil
	})
	result, err := vcli.Run(context.Background(), "echo", []string{"hello"}, vcli.WithRunner(runner))
	if err != nil {
		panic(err)
	}
	fmt.Println(result.Stdout)
	// Output: echo hello
}

func ExampleRun_withOptions() {
	runner := vcli.RunnerFunc(func(ctx context.Context, req vcli.ExecRequest) (vcli.ExecResult, error) {
		stdin, _ := io.ReadAll(req.Stdin)
		return vcli.ExecResult{
			Stdout: fmt.Sprintf("dir=%s env=%s stdin=%s", req.Dir, req.Env[0], stdin),
		}, nil
	})
	result, err := vcli.Run(context.Background(), "tool", []string{"ignored"},
		vcli.WithRunner(runner),
		vcli.WithDir("/workspace"),
		vcli.WithEnv([]string{"APP_ENV=test"}),
		vcli.WithStdin(strings.NewReader("input")),
		vcli.WithMaxOutputBytes(64),
	)
	if err != nil {
		panic(err)
	}
	fmt.Println(result.Stdout)
	// Output: dir=/workspace env=APP_ENV=test stdin=input
}

func ExampleOutput() {
	runner := vcli.RunnerFunc(func(ctx context.Context, req vcli.ExecRequest) (vcli.ExecResult, error) {
		return vcli.ExecResult{Stdout: strings.Join(req.Args, ",")}, nil
	})
	stdout, err := vcli.Output(context.Background(), "list", []string{"one", "two"}, vcli.WithRunner(runner))
	if err != nil {
		panic(err)
	}
	fmt.Println(stdout)
	// Output: one,two
}

func ExampleWithRunner() {
	runner := vcli.RunnerFunc(func(ctx context.Context, req vcli.ExecRequest) (vcli.ExecResult, error) {
		return vcli.ExecResult{Stdout: req.Name}, nil
	})
	result, err := vcli.Run(context.Background(), "fake-tool", nil, vcli.WithRunner(runner))
	if err != nil {
		panic(err)
	}
	fmt.Println(result.Stdout)
	// Output: fake-tool
}

func ExampleWithDir() {
	runner := vcli.RunnerFunc(func(ctx context.Context, req vcli.ExecRequest) (vcli.ExecResult, error) {
		return vcli.ExecResult{Stdout: req.Dir}, nil
	})
	result, err := vcli.Run(context.Background(), "pwd", nil, vcli.WithRunner(runner), vcli.WithDir("/workspace"))
	if err != nil {
		panic(err)
	}
	fmt.Println(result.Stdout)
	// Output: /workspace
}

func ExampleWithEnv() {
	runner := vcli.RunnerFunc(func(ctx context.Context, req vcli.ExecRequest) (vcli.ExecResult, error) {
		return vcli.ExecResult{Stdout: strings.Join(req.Env, ",")}, nil
	})
	result, err := vcli.Run(context.Background(), "env", nil, vcli.WithRunner(runner), vcli.WithEnv([]string{"APP_ENV=test"}))
	if err != nil {
		panic(err)
	}
	fmt.Println(result.Stdout)
	// Output: APP_ENV=test
}

func ExampleWithStdin() {
	runner := vcli.RunnerFunc(func(ctx context.Context, req vcli.ExecRequest) (vcli.ExecResult, error) {
		stdin, _ := io.ReadAll(req.Stdin)
		return vcli.ExecResult{Stdout: string(stdin)}, nil
	})
	result, err := vcli.Run(context.Background(), "cat", nil, vcli.WithRunner(runner), vcli.WithStdin(strings.NewReader("input")))
	if err != nil {
		panic(err)
	}
	fmt.Println(result.Stdout)
	// Output: input
}

func ExampleWithMaxOutputBytes() {
	runner := vcli.RunnerFunc(func(ctx context.Context, req vcli.ExecRequest) (vcli.ExecResult, error) {
		return vcli.ExecResult{Stdout: "abcdef"}, nil
	})
	result, err := vcli.Run(context.Background(), "tool", nil, vcli.WithRunner(runner), vcli.WithMaxOutputBytes(3))
	fmt.Println(result.Stdout)
	fmt.Println(errors.Is(err, vcli.ErrOutputLimitExceeded))
	// Output:
	// abc
	// true
}

func ExampleNewFlagParser() {
	parser := vcli.NewFlagParser("serve")
	port := parser.Int("port", 8080, "port to bind")
	debug := parser.Bool("debug", false, "enable debug")
	result, err := parser.Parse([]string{"--port", "9090", "--debug", "api"})
	if err != nil {
		panic(err)
	}
	fmt.Println(*port, *debug, result.Args[0])
	// Output: 9090 true api
}

func ExampleNewFlagParser_withUsageOutput() {
	var usage strings.Builder
	parser := vcli.NewFlagParser("serve", vcli.WithFlagOutput(&usage))
	_ = parser.String("config", "", "config file")
	_, err := parser.Parse([]string{"--missing"})
	fmt.Println(err != nil)
	fmt.Println(strings.Contains(usage.String(), "flag provided but not defined: -missing"))
	fmt.Println(strings.Contains(usage.String(), "-config string"))
	// Output:
	// true
	// true
	// true
}

func ExampleWithFlagOutput() {
	var usage strings.Builder
	parser := vcli.NewFlagParser("serve", vcli.WithFlagOutput(&usage))
	_ = parser.String("config", "", "config file")
	parser.Usage()
	fmt.Println(strings.Contains(usage.String(), "-config string"))
	// Output: true
}

func ExampleCommand_Execute() {
	cmd := &vcli.Command{
		Name: "hello",
		Run: func(ctx context.Context, inv *vcli.Invocation) error {
			_, _ = fmt.Fprintf(inv.Stdout, "hello %s", inv.Args[0])
			return nil
		},
	}
	var out strings.Builder
	if err := cmd.Execute(context.Background(), []string{"gopher"}, vcli.WithStdout(&out)); err != nil {
		panic(err)
	}
	fmt.Println(out.String())
	// Output: hello gopher
}

func ExampleCommand_Execute_withStderr() {
	cmd := &vcli.Command{
		Name: "serve",
		Flags: func(parser *vcli.FlagParser) {
			_ = parser.Bool("debug", false, "enable debug")
		},
		Run: func(ctx context.Context, inv *vcli.Invocation) error {
			_, _ = fmt.Fprint(inv.Stdout, "running")
			return nil
		},
	}
	var out strings.Builder
	var errOut strings.Builder
	err := cmd.Execute(context.Background(), []string{"--bad"}, vcli.WithStdout(&out), vcli.WithStderr(&errOut))
	fmt.Println(err != nil)
	fmt.Println(out.String())
	fmt.Println(strings.Contains(errOut.String(), "flag provided but not defined: -bad"))
	// Output:
	// true
	//
	// true
}

func ExampleWithStdout() {
	cmd := &vcli.Command{
		Name: "hello",
		Run: func(ctx context.Context, inv *vcli.Invocation) error {
			_, _ = fmt.Fprint(inv.Stdout, "ok")
			return nil
		},
	}
	var out strings.Builder
	if err := cmd.Execute(context.Background(), nil, vcli.WithStdout(&out)); err != nil {
		panic(err)
	}
	fmt.Println(out.String())
	// Output: ok
}

func ExampleWithStderr() {
	cmd := &vcli.Command{Name: "serve"}
	var errOut strings.Builder
	err := cmd.Execute(context.Background(), []string{"--bad"}, vcli.WithStderr(&errOut))
	fmt.Println(err != nil)
	fmt.Println(strings.Contains(errOut.String(), "flag provided but not defined: -bad"))
	// Output:
	// true
	// true
}

func ExampleRenderHelp() {
	root := &vcli.Command{Name: "app", Usage: "app <command>", Summary: "demo app"}
	root.Add(&vcli.Command{Name: "serve", Summary: "start server"})
	fmt.Print(vcli.RenderHelp(root, vcli.WithColorMode(vcli.ColorNever)))
	// Output:
	// Usage: app <command>
	//
	// demo app
	//
	// Commands:
	//   serve	start server
}

func ExampleColorize() {
	fmt.Println(vcli.Colorize("ok", vcli.Green, vcli.WithColorMode(vcli.ColorNever)))
	fmt.Println(strings.Contains(vcli.Colorize("ok", vcli.Green, vcli.WithColorMode(vcli.ColorAlways)), "\x1b["))
	// Output:
	// ok
	// true
}

func ExampleWithColorMode() {
	fmt.Println(vcli.Colorize("ok", vcli.Green, vcli.WithColorMode(vcli.ColorNever)))
	// Output: ok
}

func ExampleWithTimeout() {
	runner := vcli.RunnerFunc(func(ctx context.Context, req vcli.ExecRequest) (vcli.ExecResult, error) {
		_, hasDeadline := ctx.Deadline()
		return vcli.ExecResult{Stdout: fmt.Sprint(hasDeadline)}, nil
	})
	stdout, err := vcli.Output(
		context.Background(),
		"tool",
		nil,
		vcli.WithRunner(runner),
		vcli.WithTimeout(time.Second),
	)
	if err != nil {
		panic(err)
	}
	fmt.Println(stdout)
	// Output: true
}
