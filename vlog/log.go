package vlog

import (
	"io"
	"time"

	logx "github.com/imajinyun/knifer-go/internal/log"
)

// LogLevel is the logging level used by the logging facade.
type LogLevel = logx.Level

// Level is the logging level used by the logging facade.
type Level = logx.Level

// Log is the common logging interface.
type Log = logx.Log

// LogFactory creates loggers by name.
type LogFactory = logx.LogFactory

// LogFactoryFunc adapts a function into a LogFactory.
type LogFactoryFunc = logx.LogFactoryFunc

// ConsoleLog is the default console logger implementation.
type ConsoleLog = logx.ConsoleLog

// ConsoleColorLog is a console logger with ANSI colors.
type ConsoleColorLog = logx.ConsoleColorLog

// ConsoleLogOption customizes console logger construction.
type ConsoleLogOption = logx.ConsoleLogOption

// LoggerOption customizes logger lookup/creation for one call.
type LoggerOption = logx.LoggerOption

// ColorFactory maps log levels to ANSI colors.
type ColorFactory = logx.ColorFactory

// AbstractLog provides a reusable Log implementation skeleton.
type AbstractLog = logx.AbstractLog

const (
	// LogLevelAll enables every log level.
	LogLevelAll LogLevel = logx.LevelAll
	// LogLevelTrace is the trace level.
	LogLevelTrace LogLevel = logx.LevelTrace
	// LogLevelDebug is the debug level.
	LogLevelDebug LogLevel = logx.LevelDebug
	// LogLevelInfo is the info level.
	LogLevelInfo LogLevel = logx.LevelInfo
	// LogLevelWarn is the warn level.
	LogLevelWarn LogLevel = logx.LevelWarn
	// LogLevelError is the error level.
	LogLevelError LogLevel = logx.LevelError
	// LogLevelFatal is the fatal level.
	LogLevelFatal LogLevel = logx.LevelFatal
	// LogLevelOff disables logging.
	LogLevelOff LogLevel = logx.LevelOff
)

// NewConsoleLog creates a console logger by name.
func NewConsoleLog(name string) *ConsoleLog { return logx.NewConsoleLog(name) }

// NewConsoleLogWithOptions creates a console logger customized by options.
func NewConsoleLogWithOptions(name string, opts ...ConsoleLogOption) *ConsoleLog {
	return logx.NewConsoleLogWithOptions(name, opts...)
}

// NewConsoleColorLog creates a colored console logger by name.
func NewConsoleColorLog(name string) *ConsoleColorLog { return logx.NewConsoleColorLog(name) }

// NewConsoleColorLogWithOptions creates a colored console logger customized by options.
func NewConsoleColorLogWithOptions(name string, opts ...ConsoleLogOption) *ConsoleColorLog {
	return logx.NewConsoleColorLogWithOptions(name, opts...)
}

// WithLogTimeLayout sets the timestamp layout used by console log output.
func WithLogTimeLayout(layout string) ConsoleLogOption { return logx.WithLogTimeLayout(layout) }

// WithLogClock sets the clock used to render console log timestamps.
func WithLogClock(clock func() time.Time) ConsoleLogOption { return logx.WithLogClock(clock) }

// WithLogOutput sets the output writers used by console log output.
func WithLogOutput(out, errOut io.Writer) ConsoleLogOption { return logx.WithLogOutput(out, errOut) }

// WithLogLevel sets an instance-specific console log threshold.
func WithLogLevel(level Level) ConsoleLogOption { return logx.WithLogLevel(level) }

// WithLogColorFactory sets an instance-specific color factory for ConsoleColorLog output.
func WithLogColorFactory(f ColorFactory) ConsoleLogOption { return logx.WithLogColorFactory(f) }

// WithLoggerFactory sets the logger factory used by LoggerWithOptions or NewIsolatedLogger.
func WithLoggerFactory(factory LogFactory) LoggerOption { return logx.WithLoggerFactory(factory) }

// WithLoggerConsoleOptions builds loggers with console options for one lookup/creation call.
func WithLoggerConsoleOptions(opts ...ConsoleLogOption) LoggerOption {
	return logx.WithLoggerConsoleOptions(opts...)
}

// WithLoggerCache controls whether LoggerWithOptions may use the package-level logger cache.
func WithLoggerCache(enabled bool) LoggerOption { return logx.WithLoggerCache(enabled) }

// Logger returns a cached logger by name.
func Logger(name string) Log { return logx.Get(name) }

// LoggerWithOptions returns a logger by name with per-call factory/cache options.
func LoggerWithOptions(name string, opts ...LoggerOption) Log {
	return logx.GetWithOptions(name, opts...)
}

// NewIsolatedLogger creates a logger without reading package-level factory/cache state.
func NewIsolatedLogger(name string, opts ...LoggerOption) Log {
	return logx.NewIsolatedLogger(name, opts...)
}

// DefaultLogger returns the default logger.
func DefaultLogger() Log { return logx.GetDefault() }

// DefaultLoggerWithOptions returns the default logger with per-call factory/cache options.
func DefaultLoggerWithOptions(opts ...LoggerOption) Log { return logx.GetDefaultWithOptions(opts...) }

// SetLogFactory sets the global logging factory.
func SetLogFactory(factory LogFactory) { logx.SetFactory(factory) }

// SetLogLevel sets the console logging threshold.
func SetLogLevel(level LogLevel) { logx.SetConsoleLevel(level) }

// GetLogLevel returns the console logging threshold.
func GetLogLevel() LogLevel { return logx.GetConsoleLevel() }

// SetLogColorFactory customizes console log colors.
func SetLogColorFactory(f ColorFactory) { logx.SetColorFactory(f) }

// Trace logs trace-level output through the static logger.
func Trace(args ...any) { logx.Trace(args...) }

// TraceWithOptions logs trace-level output through a per-call logger configuration.
func TraceWithOptions(opts []LoggerOption, args ...any) { logx.TraceWithOptions(opts, args...) }

// Tracef logs formatted trace-level output through the static logger.
func Tracef(format string, args ...any) { logx.Tracef(format, args...) }

// TracefWithOptions logs formatted trace-level output through a per-call logger configuration.
func TracefWithOptions(opts []LoggerOption, format string, args ...any) {
	logx.TracefWithOptions(opts, format, args...)
}

// Debug logs debug-level output through the static logger.
func Debug(args ...any) { logx.Debug(args...) }

// DebugWithOptions logs debug-level output through a per-call logger configuration.
func DebugWithOptions(opts []LoggerOption, args ...any) { logx.DebugWithOptions(opts, args...) }

// Debugf logs formatted debug-level output through the static logger.
func Debugf(format string, args ...any) { logx.Debugf(format, args...) }

// DebugfWithOptions logs formatted debug-level output through a per-call logger configuration.
func DebugfWithOptions(opts []LoggerOption, format string, args ...any) {
	logx.DebugfWithOptions(opts, format, args...)
}

// Info logs info-level output through the static logger.
func Info(args ...any) { logx.Info(args...) }

// InfoWithOptions logs info-level output through a per-call logger configuration.
func InfoWithOptions(opts []LoggerOption, args ...any) { logx.InfoWithOptions(opts, args...) }

// Infof logs formatted info-level output through the static logger.
func Infof(format string, args ...any) { logx.Infof(format, args...) }

// InfofWithOptions logs formatted info-level output through a per-call logger configuration.
func InfofWithOptions(opts []LoggerOption, format string, args ...any) {
	logx.InfofWithOptions(opts, format, args...)
}

// Warn logs warn-level output through the static logger.
func Warn(args ...any) { logx.Warn(args...) }

// WarnWithOptions logs warn-level output through a per-call logger configuration.
func WarnWithOptions(opts []LoggerOption, args ...any) { logx.WarnWithOptions(opts, args...) }

// Warnf logs formatted warn-level output through the static logger.
func Warnf(format string, args ...any) { logx.Warnf(format, args...) }

// WarnfWithOptions logs formatted warn-level output through a per-call logger configuration.
func WarnfWithOptions(opts []LoggerOption, format string, args ...any) {
	logx.WarnfWithOptions(opts, format, args...)
}

// ErrorLog logs error-level output through the static logger.
func ErrorLog(args ...any) { logx.ErrorLog(args...) }

// ErrorLogWithOptions logs error-level output through a per-call logger configuration.
func ErrorLogWithOptions(opts []LoggerOption, args ...any) { logx.ErrorLogWithOptions(opts, args...) }

// Errorf logs formatted error-level output through the static logger.
func Errorf(format string, args ...any) { logx.Errorf(format, args...) }

// ErrorfWithOptions logs formatted error-level output through a per-call logger configuration.
func ErrorfWithOptions(opts []LoggerOption, format string, args ...any) {
	logx.ErrorfWithOptions(opts, format, args...)
}

// LogAt logs output at the provided level through the static logger.
func LogAt(level LogLevel, format string, args ...any) { logx.LogAt(level, format, args...) }

// LogAtWithOptions logs output at the provided level through a per-call logger configuration.
func LogAtWithOptions(opts []LoggerOption, level LogLevel, format string, args ...any) {
	logx.LogAtWithOptions(opts, level, format, args...)
}

// LogAtE logs output at the provided level with an error through the static logger.
func LogAtE(level LogLevel, err error, format string, args ...any) {
	logx.LogAtE(level, err, format, args...)
}

// LogAtEWithOptions logs output at the provided level with an error through a per-call logger configuration.
func LogAtEWithOptions(opts []LoggerOption, level LogLevel, err error, format string, args ...any) {
	logx.LogAtEWithOptions(opts, level, err, format, args...)
}
