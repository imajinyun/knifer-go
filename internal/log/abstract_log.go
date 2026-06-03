package log

// AbstractLog 提供 Log 接口的便捷实现，子类型只需实现 LogCore（即 LogE）即可。
// 通过将 Log 接口的方法委托给 LogE，避免在每个具体实现里重复样板代码。
type AbstractLog struct {
	// Core 由具体实现提供：在指定级别打印日志。
	// 当 IsEnabled(level) 为 false 时，AbstractLog 会跳过调用。
	Core func(level Level, err error, format string, args ...any)
	// IsEnabledFn 由具体实现提供，用于判断指定级别是否启用。
	IsEnabledFn func(level Level) bool
}

// IsEnabled 通过注入函数判断。
func (a *AbstractLog) IsEnabled(level Level) bool {
	if a.IsEnabledFn == nil {
		return true
	}
	return a.IsEnabledFn(level)
}

// IsTraceEnabled 等便捷方法。
func (a *AbstractLog) IsTraceEnabled() bool { return a.IsEnabled(LevelTrace) }
func (a *AbstractLog) IsDebugEnabled() bool { return a.IsEnabled(LevelDebug) }
func (a *AbstractLog) IsInfoEnabled() bool  { return a.IsEnabled(LevelInfo) }
func (a *AbstractLog) IsWarnEnabled() bool  { return a.IsEnabled(LevelWarn) }
func (a *AbstractLog) IsErrorEnabled() bool { return a.IsEnabled(LevelError) }

// LogE 通用入口：根据级别打印日志（带错误）。
func (a *AbstractLog) LogE(level Level, err error, format string, args ...any) {
	if !a.IsEnabled(level) {
		return
	}
	if a.Core != nil {
		a.Core(level, err, format, args...)
	}
}

// Log 通用入口：根据级别打印日志。
func (a *AbstractLog) Log(level Level, format string, args ...any) {
	a.LogE(level, nil, format, args...)
}

// Trace logs at the trace level; the following methods are level shortcuts.
func (a *AbstractLog) Trace(args ...any)                 { a.Log(LevelTrace, "", args...) }
func (a *AbstractLog) Tracef(format string, args ...any) { a.Log(LevelTrace, format, args...) }
func (a *AbstractLog) Debug(args ...any)                 { a.Log(LevelDebug, "", args...) }
func (a *AbstractLog) Debugf(format string, args ...any) { a.Log(LevelDebug, format, args...) }
func (a *AbstractLog) Info(args ...any)                  { a.Log(LevelInfo, "", args...) }
func (a *AbstractLog) Infof(format string, args ...any)  { a.Log(LevelInfo, format, args...) }
func (a *AbstractLog) Warn(args ...any)                  { a.Log(LevelWarn, "", args...) }
func (a *AbstractLog) Warnf(format string, args ...any)  { a.Log(LevelWarn, format, args...) }
func (a *AbstractLog) Error(args ...any)                 { a.Log(LevelError, "", args...) }
func (a *AbstractLog) Errorf(format string, args ...any) { a.Log(LevelError, format, args...) }
