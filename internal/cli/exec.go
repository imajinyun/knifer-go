package cli

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"slices"
	"time"
)

var (
	// ErrEmptyCommand reports that no executable name was provided.
	ErrEmptyCommand = errors.New("empty command")
	// ErrOutputLimitExceeded reports that captured stdout or stderr exceeded the configured limit.
	ErrOutputLimitExceeded = errors.New("output limit exceeded")
)

// ExecRequest describes one process execution request.
type ExecRequest struct {
	Name           string
	Args           []string
	Dir            string
	Env            []string
	Stdin          io.Reader
	MaxOutputBytes int64
}

// ExecResult describes captured process output and exit metadata.
type ExecResult struct {
	Stdout   string
	Stderr   string
	ExitCode int
	Duration time.Duration
}

// Runner executes an ExecRequest.
type Runner interface {
	Run(context.Context, ExecRequest) (ExecResult, error)
}

// RunnerFunc adapts a function to Runner.
type RunnerFunc func(context.Context, ExecRequest) (ExecResult, error)

// Run executes the runner function.
func (fn RunnerFunc) Run(ctx context.Context, req ExecRequest) (ExecResult, error) {
	return fn(ctx, req)
}

type execConfig struct {
	runner         Runner
	dir            string
	env            []string
	stdin          io.Reader
	maxOutputBytes int64
	timeout        time.Duration
}

// ExecOption customizes Run and Output.
type ExecOption func(*execConfig)

// WithRunner sets the process runner used by Run and Output.
func WithRunner(r Runner) ExecOption {
	return func(c *execConfig) {
		if r != nil {
			c.runner = r
		}
	}
}

// WithDir sets the working directory for the command.
func WithDir(dir string) ExecOption {
	return func(c *execConfig) { c.dir = dir }
}

// WithEnv appends environment entries for the command.
func WithEnv(env []string) ExecOption {
	return func(c *execConfig) { c.env = slices.Clone(env) }
}

// WithStdin sets the stdin reader for the command.
func WithStdin(r io.Reader) ExecOption {
	return func(c *execConfig) {
		if r != nil {
			c.stdin = r
		}
	}
}

// WithMaxOutputBytes limits captured stdout and stderr bytes. A non-positive value means unlimited.
func WithMaxOutputBytes(n int64) ExecOption {
	return func(c *execConfig) { c.maxOutputBytes = n }
}

// WithTimeout applies a child context timeout around command execution.
func WithTimeout(d time.Duration) ExecOption {
	return func(c *execConfig) { c.timeout = d }
}

func applyExecOptions(opts []ExecOption) execConfig {
	cfg := execConfig{runner: defaultRunner{}}
	for _, opt := range opts {
		if opt != nil {
			opt(&cfg)
		}
	}
	if cfg.runner == nil {
		cfg.runner = defaultRunner{}
	}
	cfg.env = slices.Clone(cfg.env)
	return cfg
}

// Run executes name with args using context-aware process handling.
func Run(ctx context.Context, name string, args []string, opts ...ExecOption) (ExecResult, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if name == "" {
		return ExecResult{}, ErrEmptyCommand
	}
	cfg := applyExecOptions(opts)
	if cfg.timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, cfg.timeout)
		defer cancel()
	}
	if err := ctx.Err(); err != nil {
		return ExecResult{}, err
	}
	req := ExecRequest{
		Name:           name,
		Args:           slices.Clone(args),
		Dir:            cfg.dir,
		Env:            slices.Clone(cfg.env),
		Stdin:          cfg.stdin,
		MaxOutputBytes: cfg.maxOutputBytes,
	}
	result, err := cfg.runner.Run(ctx, req)
	limited, limitErr := limitResult(result, cfg.maxOutputBytes)
	if limitErr != nil {
		return limited, limitErr
	}
	if err != nil {
		return limited, fmt.Errorf("run %q: %w", name, err)
	}
	return limited, nil
}

// Output executes name with args and returns stdout.
func Output(ctx context.Context, name string, args []string, opts ...ExecOption) (string, error) {
	result, err := Run(ctx, name, args, opts...)
	return result.Stdout, err
}

type defaultRunner struct{}

func (defaultRunner) Run(ctx context.Context, req ExecRequest) (ExecResult, error) {
	started := time.Now()
	// #nosec G204 -- vcli intentionally executes explicit command names plus args; it never invokes a shell.
	cmd := exec.CommandContext(ctx, req.Name, req.Args...)
	cmd.Dir = req.Dir
	cmd.Env = append(cmd.Environ(), req.Env...)
	cmd.Stdin = req.Stdin
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	result := ExecResult{
		Stdout:   stdout.String(),
		Stderr:   stderr.String(),
		ExitCode: exitCode(err),
		Duration: time.Since(started),
	}
	limited, limitErr := limitResult(result, req.MaxOutputBytes)
	if limitErr != nil {
		return limited, limitErr
	}
	if err != nil {
		return limited, err
	}
	return limited, nil
}

func exitCode(err error) int {
	if err == nil {
		return 0
	}
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		return exitErr.ExitCode()
	}
	return -1
}

func limitResult(result ExecResult, max int64) (ExecResult, error) {
	if max <= 0 {
		return result, nil
	}
	stdout := []byte(result.Stdout)
	stderr := []byte(result.Stderr)
	if int64(len(stdout)) > max {
		result.Stdout = string(stdout[:max])
		result.Stderr = ""
		return result, ErrOutputLimitExceeded
	}
	remaining := max - int64(len(stdout))
	if int64(len(stderr)) > remaining {
		result.Stderr = string(stderr[:remaining])
		return result, ErrOutputLimitExceeded
	}
	return result, nil
}
