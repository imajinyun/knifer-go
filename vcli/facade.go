package vcli

import (
	"context"

	"github.com/imajinyun/knifer-go/internal/cli"
)

// Run executes name with args using context-aware process handling.
func Run(ctx context.Context, name string, args []string, opts ...ExecOption) (ExecResult, error) {
	return cli.Run(ctx, name, args, opts...)
}

// Output executes name with args and returns stdout.
func Output(ctx context.Context, name string, args []string, opts ...ExecOption) (string, error) {
	return cli.Output(ctx, name, args, opts...)
}

// NewFlagParser creates a flag parser for one command.
func NewFlagParser(name string, opts ...FlagParserOption) *FlagParser {
	return cli.NewFlagParser(name, opts...)
}

// RenderHelp returns deterministic help text for cmd.
func RenderHelp(cmd *Command, opts ...ColorOption) string {
	return cli.RenderHelp(cmd, opts...)
}

// Colorize wraps text in ANSI escape codes when enabled.
func Colorize(text string, color Color, opts ...ColorOption) string {
	return cli.Colorize(text, color, opts...)
}
