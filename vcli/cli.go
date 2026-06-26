package vcli

import (
	"context"
	"io"
	"time"

	"github.com/imajinyun/knifer-go/internal/cli"
)

// ExecRequest describes one process execution request.
type ExecRequest = cli.ExecRequest

// ExecResult describes captured process output and exit metadata.
type ExecResult = cli.ExecResult

// Runner executes an ExecRequest.
type Runner = cli.Runner

// RunnerFunc adapts a function to Runner.
type RunnerFunc = cli.RunnerFunc

// ExecOption customizes Run and Output.
type ExecOption = cli.ExecOption

// FlagParseResult contains the non-flag arguments left after parsing.
type FlagParseResult = cli.FlagParseResult

// FlagParser wraps flag parsing with deterministic error handling.
type FlagParser = cli.FlagParser

// FlagParserOption customizes a FlagParser.
type FlagParserOption = cli.FlagParserOption

// Handler runs a command invocation.
type Handler = cli.Handler

// Command describes a lightweight command or subcommand.
type Command = cli.Command

// Invocation contains parsed arguments and command I/O streams.
type Invocation = cli.Invocation

// ExecuteOption customizes command execution.
type ExecuteOption = cli.ExecuteOption

// ColorMode controls ANSI color output.
type ColorMode = cli.ColorMode

const (
	// ColorAuto enables color output for callers that opt into automatic behavior.
	ColorAuto = cli.ColorAuto
	// ColorAlways always emits ANSI color escape sequences.
	ColorAlways = cli.ColorAlways
	// ColorNever never emits ANSI color escape sequences.
	ColorNever = cli.ColorNever
)

// Color names supported ANSI foreground colors.
type Color = cli.Color

const (
	// Red is the ANSI red foreground color.
	Red = cli.Red
	// Green is the ANSI green foreground color.
	Green = cli.Green
	// Yellow is the ANSI yellow foreground color.
	Yellow = cli.Yellow
	// Blue is the ANSI blue foreground color.
	Blue = cli.Blue
	// Bold is the ANSI bold text attribute.
	Bold = cli.Bold
)

// ColorOption customizes color rendering.
type ColorOption = cli.ColorOption

var (
	// ErrEmptyCommand reports that no executable name was provided.
	ErrEmptyCommand = cli.ErrEmptyCommand
	// ErrOutputLimitExceeded reports that captured stdout or stderr exceeded the configured limit.
	ErrOutputLimitExceeded = cli.ErrOutputLimitExceeded
	// ErrUsage reports invalid CLI arguments or command selection.
	ErrUsage = cli.ErrUsage
)

// WithRunner sets the process runner used by Run and Output.
func WithRunner(r Runner) ExecOption { return cli.WithRunner(r) }

// WithDir sets the working directory for the command.
func WithDir(dir string) ExecOption { return cli.WithDir(dir) }

// WithEnv appends environment entries for the command.
func WithEnv(env []string) ExecOption { return cli.WithEnv(env) }

// WithStdin sets the stdin reader for the command.
func WithStdin(r io.Reader) ExecOption { return cli.WithStdin(r) }

// WithMaxOutputBytes limits captured stdout and stderr bytes. A non-positive value means unlimited.
func WithMaxOutputBytes(n int64) ExecOption { return cli.WithMaxOutputBytes(n) }

// WithTimeout applies a child context timeout around command execution.
func WithTimeout(d time.Duration) ExecOption { return cli.WithTimeout(d) }

// WithFlagOutput sets where parser usage text is written.
func WithFlagOutput(w io.Writer) FlagParserOption { return cli.WithFlagOutput(w) }

// WithStdout sets command stdout.
func WithStdout(w io.Writer) ExecuteOption { return cli.WithStdout(w) }

// WithStderr sets command stderr.
func WithStderr(w io.Writer) ExecuteOption { return cli.WithStderr(w) }

// WithColorMode sets ANSI color behavior.
func WithColorMode(mode ColorMode) ColorOption { return cli.WithColorMode(mode) }

// Compile-time check that RunnerFunc keeps the public context-aware signature.
var _ Runner = RunnerFunc(func(context.Context, ExecRequest) (ExecResult, error) { return ExecResult{}, nil })
