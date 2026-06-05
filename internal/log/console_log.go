package log

import (
	"fmt"
	"io"
	"os"
	"sync"
	"time"
)

// ConsoleLog 对应 the utility ConsoleLog，使用标准输出/错误打印日志。
//
// 默认级别为 LevelDebug，可通过 SetConsoleLevel 全局调整。
// 输出格式为：[date] [level] name: msg
type ConsoleLog struct {
	*AbstractLog
	name string
	// out / errOut 可由测试注入；为 nil 时使用 os.Stdout / os.Stderr。
	out    io.Writer
	errOut io.Writer
	// clock / timeLayout 可由测试注入；为空时使用 time.Now / 默认布局。
	clock      func() time.Time
	timeLayout string
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
		c.SetOutput(out, errOut)
	}
}

var (
	consoleLevelMu sync.RWMutex
	consoleLevel   = LevelDebug
)

// SetConsoleLevel 全局设置控制台日志级别（小于该级别的日志将被过滤）。
func SetConsoleLevel(level Level) {
	consoleLevelMu.Lock()
	defer consoleLevelMu.Unlock()
	consoleLevel = level
}

// GetConsoleLevel 返回当前控制台日志级别。
func GetConsoleLevel() Level {
	consoleLevelMu.RLock()
	defer consoleLevelMu.RUnlock()
	return consoleLevel
}

// NewConsoleLog 创建一个使用控制台输出的 Log 实例。
func NewConsoleLog(name string) *ConsoleLog {
	return NewConsoleLogWithOptions(name)
}

// NewConsoleLogWithOptions 创建一个使用控制台输出的 Log 实例，并应用构造选项。
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
		IsEnabledFn: func(level Level) bool { return GetConsoleLevel() <= level },
	}
	return c
}

// GetName 返回日志名称。
func (c *ConsoleLog) GetName() string { return c.name }

// SetOutput 设置标准输出目标，主要用于测试。
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

// write 是底层写入逻辑，由 AbstractLog.Core 调用。
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
