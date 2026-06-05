package log

import (
	"fmt"
	"sync"
)

// ANSI 颜色码。
const (
	ansiReset   = "\033[0m"
	ansiDefault = "\033[39m"
	ansiBlack   = "\033[30m"
	ansiRed     = "\033[31m"
	ansiGreen   = "\033[32m"
	ansiYellow  = "\033[33m"
	ansiBlue    = "\033[34m"
	ansiMagenta = "\033[35m"
	ansiCyan    = "\033[36m"
	ansiWhite   = "\033[37m"
)

// ColorFactory 根据级别返回对应的 ANSI 颜色码。
type ColorFactory func(level Level) string

var (
	colorFactoryMu sync.RWMutex
	colorFactory   ColorFactory = defaultColorFactory
)

func defaultColorFactory(level Level) string {
	switch level {
	case LevelDebug, LevelInfo:
		return ansiGreen
	case LevelWarn:
		return ansiYellow
	case LevelError, LevelFatal:
		return ansiRed
	case LevelTrace:
		return ansiMagenta
	default:
		return ansiDefault
	}
}

// SetColorFactory 自定义颜色工厂。
func SetColorFactory(f ColorFactory) {
	if f == nil {
		return
	}
	colorFactoryMu.Lock()
	defer colorFactoryMu.Unlock()
	colorFactory = f
}

// getColorFactory 取当前颜色工厂。
func getColorFactory() ColorFactory {
	colorFactoryMu.RLock()
	defer colorFactoryMu.RUnlock()
	return colorFactory
}

// ConsoleColorLog 对应 the utility ConsoleColorLog，使用 ANSI 颜色打印日志。
type ConsoleColorLog struct {
	*ConsoleLog
}

// NewConsoleColorLog 创建一个带颜色的控制台日志实例。
func NewConsoleColorLog(name string) *ConsoleColorLog {
	return NewConsoleColorLogWithOptions(name)
}

// NewConsoleColorLogWithOptions 创建一个带颜色的控制台日志实例，并应用构造选项。
func NewConsoleColorLogWithOptions(name string, opts ...ConsoleLogOption) *ConsoleColorLog {
	base := NewConsoleLogWithOptions(name, opts...)
	c := &ConsoleColorLog{ConsoleLog: base}
	// 替换 Core 为彩色实现。
	base.Core = c.write
	return c
}

// write 带颜色的写入实现。
func (c *ConsoleColorLog) write(level Level, err error, format string, args ...any) {
	msg := renderLogMessage(format, args...)
	color := getColorFactory()(level)
	line := fmt.Sprintf(
		"%s[%s]%s %s[%-5s]%s %s%s%s: %s",
		ansiWhite, c.now().Format(c.layout()), ansiReset,
		color, level.String(), ansiReset,
		ansiCyan, c.name, ansiReset,
		msg,
	)
	if err != nil {
		line = line + " | error: " + err.Error()
	}
	w := c.targetWriter(level)
	_, _ = fmt.Fprintln(w, line)
}
