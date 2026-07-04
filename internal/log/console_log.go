package log

import (
	"fmt"
	"io"
	"os"
	"sync"
	"time"
)

// ConsoleLog matches the utility ConsoleLog and prints logs to stdout or stderr.
//
// The default level is LevelDebug and can be adjusted globally with SetConsoleLevel.
// Output format: [date] [level] name: msg
type ConsoleLog struct {
	*AbstractLog
	name string
	// out / errOut can be injected by tests; nil uses os.Stdout or os.Stderr.
	out    io.Writer
	errOut io.Writer
	// clock / timeLayout can be injected by tests; empty values use time.Now or the default layout.
	clock      func() time.Time
	timeLayout string
	level      *Level
	// colorFactory is used by ConsoleColorLog instances; nil falls back to the package default.
	colorFactory ColorFactory
}

const consoleLogTimeLayout = "2006-01-02 15:04:05"

// ConsoleLogOption customizes console logger construction.
type ConsoleLogOption func(*ConsoleLog)

// WithLogTimeLayout sets the timestamp layout used by console log output.
func WithLogTimeLayout(layout string) ConsoleLogOption {
	return func(c *ConsoleLog) {
		if layout != "" {
			c.timeLayout = layout
		}
	}
}

// WithLogClock sets the clock used to render console log timestamps.
func WithLogClock(clock func() time.Time) ConsoleLogOption {
	return func(c *ConsoleLog) {
		if clock != nil {
			c.clock = clock
		}
	}
}

// WithLogOutput sets the output writers used by console log output.
func WithLogOutput(out, errOut io.Writer) ConsoleLogOption {
	return func(c *ConsoleLog) {
		if out != nil {
			c.out = out
		}
		if errOut != nil {
			c.errOut = errOut
		}
	}
}

// WithLogLevel sets an instance-specific console log threshold.
func WithLogLevel(level Level) ConsoleLogOption {
	return func(c *ConsoleLog) {
		c.level = &level
	}
}

// WithLogColorFactory sets an instance-specific color factory for ConsoleColorLog output.
func WithLogColorFactory(f ColorFactory) ConsoleLogOption {
	return func(c *ConsoleLog) {
		if f != nil {
			c.colorFactory = f
		}
	}
}

var (
	consoleLevelMu sync.RWMutex
	consoleLevel   = LevelDebug
)

// SetConsoleLevel sets the global console log level; logs below this level are filtered.
func SetConsoleLevel(level Level) {
	consoleLevelMu.Lock()
	defer consoleLevelMu.Unlock()
	consoleLevel = level
}

// GetConsoleLevel returns the current console log level.
func GetConsoleLevel() Level {
	consoleLevelMu.RLock()
	defer consoleLevelMu.RUnlock()
	return consoleLevel
}

// NewConsoleLog creates a Log instance that writes to the console.
func NewConsoleLog(name string) *ConsoleLog {
	return NewConsoleLogWithOptions(name)
}

// NewConsoleLogWithOptions creates a Log instance that writes to the console and applies constructor options.
func NewConsoleLogWithOptions(name string, opts ...ConsoleLogOption) *ConsoleLog {
	c := &ConsoleLog{
		name:       name,
		clock:      time.Now,
		timeLayout: consoleLogTimeLayout,
	}
	for _, opt := range opts {
		if opt != nil {
			opt(c)
		}
	}
	c.AbstractLog = &AbstractLog{
		Core:        c.write,
		IsEnabledFn: c.isEnabled,
	}
	return c
}

// GetName returns the log name.
func (c *ConsoleLog) GetName() string { return c.name }

// SetOutput sets the stdout target, mainly for tests.
func (c *ConsoleLog) SetOutput(out, errOut io.Writer) {
	c.out = out
	c.errOut = errOut
}

func (c *ConsoleLog) now() time.Time {
	if c.clock != nil {
		return c.clock()
	}
	return time.Now()
}

func (c *ConsoleLog) layout() string {
	if c.timeLayout != "" {
		return c.timeLayout
	}
	return consoleLogTimeLayout
}

func (c *ConsoleLog) isEnabled(level Level) bool {
	if c.level != nil {
		return *c.level <= level
	}
	return GetConsoleLevel() <= level
}

// write is the low-level write logic called by AbstractLog.Core.
func (c *ConsoleLog) write(level Level, err error, format string, args ...any) {
	msg := renderLogMessage(format, args...)
	line := fmt.Sprintf("[%s] [%-5s] %s: %s", c.now().Format(c.layout()), level.String(), c.name, msg)
	if err != nil {
		line = line + " | error: " + err.Error()
	}
	w := c.targetWriter(level)
	_, _ = fmt.Fprintln(w, line)
}

func (c *ConsoleLog) targetWriter(level Level) io.Writer {
	if level >= LevelWarn {
		if c.errOut != nil {
			return c.errOut
		}
		return os.Stderr
	}
	if c.out != nil {
		return c.out
	}
	return os.Stdout
}
