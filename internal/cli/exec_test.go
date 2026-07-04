package cli

import (
	"context"
	"errors"
	"io"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"
)

type fakeRunner struct {
	requests []ExecRequest
	result   ExecResult
	err      error
}

func (r *fakeRunner) Run(ctx context.Context, req ExecRequest) (ExecResult, error) {
	r.requests = append(r.requests, req)
	select {
	case <-ctx.Done():
		return ExecResult{}, ctx.Err()
	default:
	}
	return r.result, r.err
}

func TestRunUsesInjectedRunnerAndClonesMutableInputs(t *testing.T) {
	runner := &fakeRunner{result: ExecResult{Stdout: "ok", ExitCode: 0}}
	env := []string{"A=1"}
	result, err := Run(
		context.Background(),
		"git",
		[]string{"status"},
		WithRunner(runner),
		WithDir("/tmp/work"),
		WithEnv(env),
		WithStdin(strings.NewReader("input")),
	)
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if result.Stdout != "ok" || result.ExitCode != 0 {
		t.Fatalf("Run result = %+v", result)
	}
	env[0] = "A=changed"
	if len(runner.requests) != 1 {
		t.Fatalf("runner requests = %d", len(runner.requests))
	}
	got := runner.requests[0]
	if got.Name != "git" || !reflect.DeepEqual(got.Args, []string{"status"}) || got.Dir != "/tmp/work" {
		t.Fatalf("request = %+v", got)
	}
	if !reflect.DeepEqual(got.Env, []string{"A=1"}) {
		t.Fatalf("request env was not cloned: %#v", got.Env)
	}
	if got.Stdin == nil {
		t.Fatalf("request stdin is nil")
	}
}

func TestNilStdinOptionDoesNotClearPreviousReader(t *testing.T) {
	runner := &fakeRunner{result: ExecResult{ExitCode: 0}}
	result, err := Run(
		context.Background(),
		"tool",
		nil,
		WithRunner(runner),
		WithStdin(strings.NewReader("input")),
		WithStdin(nil),
	)
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if result.ExitCode != 0 {
		t.Fatalf("Run result = %+v", result)
	}
	if len(runner.requests) != 1 {
		t.Fatalf("runner requests = %d", len(runner.requests))
	}
	if runner.requests[0].Stdin == nil {
		t.Fatal("nil WithStdin cleared the previously configured reader")
	}
}

func TestRunRejectsEmptyCommandName(t *testing.T) {
	_, err := Run(context.Background(), "", nil)
	if !errors.Is(err, ErrEmptyCommand) {
		t.Fatalf("Run empty command error = %v, want ErrEmptyCommand", err)
	}
}

func TestRunPropagatesContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := Run(ctx, "tool", nil, WithRunner(&fakeRunner{}))
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("Run canceled error = %v, want context.Canceled", err)
	}
}

func TestRunWrapsRunnerError(t *testing.T) {
	runnerErr := errors.New("boom")
	_, err := Run(context.Background(), "tool", nil, WithRunner(&fakeRunner{err: runnerErr}))
	if !errors.Is(err, runnerErr) {
		t.Fatalf("Run error = %v, want wrapped runner error", err)
	}
}

func TestRunEnforcesOutputLimitOnInjectedRunner(t *testing.T) {
	runner := &fakeRunner{result: ExecResult{Stdout: "abcdef", Stderr: "xyz"}}
	result, err := Run(context.Background(), "tool", nil, WithRunner(runner), WithMaxOutputBytes(5))
	if !errors.Is(err, ErrOutputLimitExceeded) {
		t.Fatalf("Run limit error = %v, want ErrOutputLimitExceeded", err)
	}
	if result.Stdout != "abcde" || result.Stderr != "" {
		t.Fatalf("limited result = %+v", result)
	}
}

func TestRunEnforcesOutputLimitAcrossStdoutAndStderr(t *testing.T) {
	runner := &fakeRunner{result: ExecResult{Stdout: "abc", Stderr: "def"}}
	result, err := Run(context.Background(), "tool", nil, WithRunner(runner), WithMaxOutputBytes(5))
	if !errors.Is(err, ErrOutputLimitExceeded) {
		t.Fatalf("Run limit error = %v, want ErrOutputLimitExceeded", err)
	}
	if result.Stdout != "abc" || result.Stderr != "de" {
		t.Fatalf("limited result = %+v", result)
	}
}

func TestOutputReturnsStdout(t *testing.T) {
	runner := &fakeRunner{result: ExecResult{Stdout: "hello\n", ExitCode: 0}}
	got, err := Output(context.Background(), "printf", []string{"hello"}, WithRunner(runner))
	if err != nil {
		t.Fatalf("Output returned error: %v", err)
	}
	if got != "hello\n" {
		t.Fatalf("Output = %q", got)
	}
}

func TestWithTimeoutAppliesDeadline(t *testing.T) {
	runner := RunnerFunc(func(ctx context.Context, req ExecRequest) (ExecResult, error) {
		deadline, ok := ctx.Deadline()
		if !ok {
			t.Fatalf("ctx has no deadline")
		}
		if time.Until(deadline) > time.Second {
			t.Fatalf("deadline too far away: %v", deadline)
		}
		return ExecResult{Stdout: "ok"}, nil
	})
	_, err := Run(context.Background(), "tool", nil, WithRunner(runner), WithTimeout(time.Second))
	if err != nil {
		t.Fatalf("Run with timeout returned error: %v", err)
	}
}

func TestDefaultRunnerCapturesSuccessfulProcess(t *testing.T) {
	exe := testExecutable(t)
	result, err := Run(
		context.Background(),
		exe,
		[]string{"-test.run=TestDefaultRunnerHelperProcess", "--", "success"},
		WithEnv([]string{"GO_KNIFER_CLI_HELPER=1"}),
	)
	if err != nil {
		t.Fatalf("Run helper success returned error: %v", err)
	}
	if result.Stdout != "stdout" || result.Stderr != "stderr" || result.ExitCode != 0 {
		t.Fatalf("result = %+v", result)
	}
	if result.Duration <= 0 {
		t.Fatalf("duration = %s, want positive duration", result.Duration)
	}
}

func TestDefaultRunnerCapturesExitCode(t *testing.T) {
	exe := testExecutable(t)
	result, err := Run(
		context.Background(),
		exe,
		[]string{"-test.run=TestDefaultRunnerHelperProcess", "--", "fail"},
		WithEnv([]string{"GO_KNIFER_CLI_HELPER=1"}),
	)
	if err == nil {
		t.Fatal("Run helper fail returned nil error")
	}
	if result.ExitCode != 7 || result.Stderr != "failed" {
		t.Fatalf("result = %+v, want exit code 7 and stderr", result)
	}
}

func TestDefaultRunnerAppliesOutputLimit(t *testing.T) {
	exe := testExecutable(t)
	result, err := Run(
		context.Background(),
		exe,
		[]string{"-test.run=TestDefaultRunnerHelperProcess", "--", "large"},
		WithEnv([]string{"GO_KNIFER_CLI_HELPER=1"}),
		WithMaxOutputBytes(4),
	)
	if !errors.Is(err, ErrOutputLimitExceeded) {
		t.Fatalf("Run helper large error = %v, want ErrOutputLimitExceeded", err)
	}
	if result.Stdout != "abcd" || result.Stderr != "" {
		t.Fatalf("limited result = %+v", result)
	}
}

func TestDefaultRunnerHelperProcess(t *testing.T) {
	if os.Getenv("GO_KNIFER_CLI_HELPER") != "1" {
		return
	}
	args := os.Args
	for len(args) > 0 && args[0] != "--" {
		args = args[1:]
	}
	if len(args) < 2 {
		os.Exit(2)
	}
	switch args[1] {
	case "success":
		_, _ = os.Stdout.WriteString("stdout")
		_, _ = os.Stderr.WriteString("stderr")
	case "fail":
		_, _ = os.Stderr.WriteString("failed")
		os.Exit(7)
	case "large":
		_, _ = os.Stdout.WriteString("abcdef")
	default:
		os.Exit(3)
	}
	os.Exit(0)
}

func testExecutable(t *testing.T) string {
	t.Helper()
	exe, err := os.Executable()
	if err != nil {
		t.Fatalf("os.Executable: %v", err)
	}
	return exe
}

func TestExecRequestStdinAcceptsNilAndReader(t *testing.T) {
	var _ io.Reader = strings.NewReader("x")
	request := ExecRequest{Name: "tool", Stdin: strings.NewReader("x")}
	if request.Stdin == nil {
		t.Fatalf("stdin reader is nil")
	}
}
