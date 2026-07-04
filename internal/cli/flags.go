package cli

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"time"
)

// ErrUsage reports invalid CLI arguments or command selection.
var ErrUsage = errors.New("usage error")

// FlagParseResult contains the non-flag arguments left after parsing.
type FlagParseResult struct {
	Args []string
}

type flagParserConfig struct {
	output io.Writer
}

// FlagParserOption customizes a FlagParser.
type FlagParserOption func(*flagParserConfig)

// WithFlagOutput sets where parser usage text is written.
func WithFlagOutput(w io.Writer) FlagParserOption {
	return func(c *flagParserConfig) {
		if w != nil {
			c.output = w
		}
	}
}

// FlagParser wraps flag.FlagSet with deterministic error handling.
type FlagParser struct {
	set *flag.FlagSet
}

// NewFlagParser creates a flag parser for one command.
func NewFlagParser(name string, opts ...FlagParserOption) *FlagParser {
	cfg := flagParserConfig{output: io.Discard}
	for _, opt := range opts {
		if opt != nil {
			opt(&cfg)
		}
	}
	set := flag.NewFlagSet(name, flag.ContinueOnError)
	set.SetOutput(cfg.output)
	return &FlagParser{set: set}
}

// String defines a string flag.
func (p *FlagParser) String(name, value, usage string) *string {
	return p.set.String(name, value, usage)
}

// Int defines an int flag.
func (p *FlagParser) Int(name string, value int, usage string) *int {
	return p.set.Int(name, value, usage)
}

// Bool defines a bool flag.
func (p *FlagParser) Bool(name string, value bool, usage string) *bool {
	return p.set.Bool(name, value, usage)
}

// Duration defines a time.Duration flag.
func (p *FlagParser) Duration(name string, value time.Duration, usage string) *time.Duration {
	return p.set.Duration(name, value, usage)
}

// Parse parses args and returns positional arguments.
func (p *FlagParser) Parse(args []string) (FlagParseResult, error) {
	if err := p.set.Parse(args); err != nil {
		return FlagParseResult{}, fmt.Errorf("parse flags: %w", errors.Join(ErrUsage, err))
	}
	return FlagParseResult{Args: append([]string(nil), p.set.Args()...)}, nil
}

// Usage writes parser usage text to the configured output.
func (p *FlagParser) Usage() {
	p.set.PrintDefaults()
}
