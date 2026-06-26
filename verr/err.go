package verr

import (
	"context"
	"io"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/sirupsen/logrus"

	errimpl "github.com/imajinyun/knifer-go/internal/errx"
)

const (
	// SentryDSN is the environment variable used to override the configured DSN.
	SentryDSN = errimpl.SentryDSN
)

// EmptyFormatter suppresses logrus output while still allowing hooks to run.
var EmptyFormatter = errimpl.EmptyFormatter

// WithStack is implemented by errors that can expose a string stack trace.
type WithStack = errimpl.WithStack

// PanicError wraps a recovered panic value and records the recovery stack.
type PanicError = errimpl.PanicError

// WithStackTrace is implemented by errors that expose structured stack frames.
type WithStackTrace = errimpl.WithStackTrace

// Frame represents a program counter inside a stack frame.
type Frame = errimpl.Frame

// StackTrace is a stack of frames from innermost to outermost.
type StackTrace = errimpl.StackTrace

// StackTraceOption customizes stack trace capture.
type StackTraceOption = errimpl.StackTraceOption

// StackOption customizes fallback stack capture.
type StackOption = errimpl.StackOption

// LogFunc writes an error log entry for errx helpers.
type LogFunc = errimpl.LogFunc

// DebugStackFunc captures the current goroutine stack.
type DebugStackFunc = errimpl.DebugStackFunc

// CallersFunc captures call stack PCs.
type CallersFunc = errimpl.CallersFunc

// FuncForPCFunc resolves a PC into file, line, and function name.
type FuncForPCFunc = errimpl.FuncForPCFunc

// ExitOption customizes MustExitWithOptions.
type ExitOption = errimpl.ExitOption

// InitOption customizes logrus/Sentry initialization.
type InitOption = errimpl.InitOption

// Collector runs functions, recovers panics, logs failures, and aggregates errors.
type Collector = errimpl.Collector

// Timer stops a wait timer created by TimerFactory.
type Timer = errimpl.Timer

// TimerFactory creates a timer channel and stopper for Collector waits.
type TimerFactory = errimpl.TimerFactory

// WaitOption customizes a single Collector wait call.
type WaitOption = errimpl.WaitOption

// CollectorOption customizes Collector construction.
type CollectorOption = errimpl.CollectorOption

// Wrapper executes a function with panic recovery and optional logging.
type Wrapper = errimpl.Wrapper

// NewCollector creates a Collector that logs failures at error level.
func NewCollector() *Collector { return NewCollectorWithOptions() }

// NewCollectorWithOptions creates a Collector customized by options.
func NewCollectorWithOptions(opts ...CollectorOption) *Collector {
	return errimpl.NewCollectorWithOptions(opts...)
}

// GetStack returns the stack attached to err, or the current goroutine stack.
func GetStack(err error) string { return GetStackWithOptions(err) }

// ErrorIs is like errors.Is, but it also checks each member of a multierror.
func ErrorIs(err error, target error) bool { return errimpl.ErrorIs(err, target) }

// MustExit logs err and panics when err is non-nil.
func MustExit(ctx context.Context, err error) { MustExitWithOptions(ctx, err) }

// MustExitWithOptions logs err and panics when err is non-nil with custom options.
func MustExitWithOptions(ctx context.Context, err error, opts ...ExitOption) {
	errimpl.MustExitWithOptions(ctx, err, opts...)
}

// Init configures logrus output and optional Sentry forwarding.
func Init(sentryDSN string) { InitWithOptions(WithSentryDSN(sentryDSN)) }

// WithSentryDSN sets the DSN used for Sentry forwarding.
func WithSentryDSN(dsn string) InitOption { return errimpl.WithSentryDSN(dsn) }

// WithSentryEnvKey sets the environment variable key used to resolve the Sentry DSN.
func WithSentryEnvKey(key string) InitOption { return errimpl.WithSentryEnvKey(key) }

// WithLogOutput sets the log output writer.
func WithLogOutput(output io.Writer) InitOption { return errimpl.WithLogOutput(output) }

// WithLogFormatter sets the logrus formatter used by InitWithOptions.
func WithLogFormatter(formatter logrus.Formatter) InitOption {
	return errimpl.WithLogFormatter(formatter)
}

// WithReportCaller controls whether logrus reports caller information.
func WithReportCaller(reportCaller bool) InitOption { return errimpl.WithReportCaller(reportCaller) }

// WithSentryLevels sets which log levels are forwarded to Sentry.
func WithSentryLevels(levels ...logrus.Level) InitOption { return errimpl.WithSentryLevels(levels...) }

// WithEnvLookupFunc sets the environment lookup used to override the Sentry DSN.
func WithEnvLookupFunc(getenv func(string) string) InitOption {
	return errimpl.WithEnvLookupFunc(getenv)
}

// WithSentryClientOptions sets sentry-go client options used when creating the Sentry client.
func WithSentryClientOptions(options sentry.ClientOptions) InitOption {
	return errimpl.WithSentryClientOptions(options)
}

// WithSentryClient sets the sentry-go client passed to the Sentry hook factory.
func WithSentryClient(client *sentry.Client) InitOption { return errimpl.WithSentryClient(client) }

// WithSentryClientFactory sets the factory used to create sentry-go clients.
func WithSentryClientFactory(factory func(sentry.ClientOptions) (*sentry.Client, error)) InitOption {
	return errimpl.WithSentryClientFactory(factory)
}

// WithSentryHookFactory sets the factory used to create the Sentry logrus hook.
func WithSentryHookFactory(factory func(*sentry.Client, []logrus.Level) (logrus.Hook, error)) InitOption {
	return errimpl.WithSentryHookFactory(factory)
}

// WithLogHookAdder sets the function used to register the Sentry hook.
func WithLogHookAdder(addHook func(logrus.Hook)) InitOption {
	return errimpl.WithLogHookAdder(addHook)
}

// WithLogrusConfigurer sets the logrus global configuration functions used during initialization.
func WithLogrusConfigurer(setReportCaller func(bool), setOutput func(io.Writer), setFormatter func(logrus.Formatter)) InitOption {
	return errimpl.WithLogrusConfigurer(setReportCaller, setOutput, setFormatter)
}

// WithInitErrorLogger sets the logger used for initialization failures.
func WithInitErrorLogger(logError func(error, string)) InitOption {
	return errimpl.WithInitErrorLogger(logError)
}

// InitWithOptions configures logrus output and optional Sentry forwarding with options.
func InitWithOptions(opts ...InitOption) { errimpl.InitWithOptions(opts...) }

// NewIsolatedLogrusWithOptions creates a standalone logrus logger without mutating global logrus/Sentry state.
func NewIsolatedLogrusWithOptions(opts ...InitOption) *logrus.Logger {
	return errimpl.NewIsolatedLogrusWithOptions(opts...)
}

// Wrap creates a recoverable function wrapper.
func Wrap(f func() error) *Wrapper { return errimpl.Wrap(f) }

// Recover executes f with panic recovery and logs failures at error level.
func Recover(f func() error, format string, args ...any) error {
	return errimpl.Recover(f, format, args...)
}

// RecoverWithoutError executes f with panic recovery and logs failures at error level.
func RecoverWithoutError(f func(), format string, args ...any) error {
	return errimpl.RecoverWithoutError(f, format, args...)
}

// GetStackTrace captures the current goroutine stack trace.
func GetStackTrace(skip int) StackTrace { return GetStackTraceWithOptions(WithStackSkip(skip)) }

// WithStackSkip sets how many caller frames to skip while capturing a stack trace.
func WithStackSkip(skip int) StackTraceOption { return errimpl.WithStackSkip(skip) }

// WithStackDepth limits the number of captured stack frames.
func WithStackDepth(depth int) StackTraceOption { return errimpl.WithStackDepth(depth) }

// WithCallersFunc sets the function used to capture stack PCs.
func WithCallersFunc(callers CallersFunc) StackTraceOption { return errimpl.WithCallersFunc(callers) }

// WithFuncForPCFunc sets the resolver used to format captured stack frames.
func WithFuncForPCFunc(fn FuncForPCFunc) StackTraceOption { return errimpl.WithFuncForPCFunc(fn) }

// WithStackFrameCache controls whether captured frame metadata is stored in the package-level cache.
func WithStackFrameCache(enabled bool) StackTraceOption { return errimpl.WithStackFrameCache(enabled) }

// ResetStackFrameCache clears cached stack frame metadata captured by GetStackTraceWithOptions.
func ResetStackFrameCache() { errimpl.ResetStackFrameCache() }

// WithDebugStackFunc sets the function used when err does not carry a stack.
func WithDebugStackFunc(fn DebugStackFunc) StackOption { return errimpl.WithDebugStackFunc(fn) }

// GetStackWithOptions returns the stack attached to err, or captures a stack with options.
func GetStackWithOptions(err error, opts ...StackOption) string {
	return errimpl.GetStackWithOptions(err, opts...)
}

// WithExitLogFunc sets the logger used by MustExitWithOptions.
func WithExitLogFunc(logFunc LogFunc) ExitOption { return errimpl.WithExitLogFunc(logFunc) }

// WithExitPanicFunc sets the panic function used by MustExitWithOptions.
func WithExitPanicFunc(panicFunc func(error)) ExitOption { return errimpl.WithExitPanicFunc(panicFunc) }

// ConfigureDefaultLogFunc sets the package-level logger used by errx helpers when no logger is provided.
func ConfigureDefaultLogFunc(logFunc LogFunc) { errimpl.ConfigureDefaultLogFunc(logFunc) }

// ResetDefaultLogFunc restores the package-level logger to the logrus-backed default.
func ResetDefaultLogFunc() { errimpl.ResetDefaultLogFunc() }

// GetStackTraceWithOptions captures the current goroutine stack trace with options.
func GetStackTraceWithOptions(opts ...StackTraceOption) StackTrace {
	return errimpl.GetStackTraceWithOptions(opts...)
}

// WithLevel sets the log level used for recovered or returned errors.
func WithLevel(c *Collector, level logrus.Level) *Collector { return c.WithLevel(level) }

// WithTimerFactory sets the default timer factory used by Collector.WaitUntil.
func WithTimerFactory(c *Collector, factory TimerFactory) *Collector {
	return c.WithTimerFactory(factory)
}

// WithLogFunc sets the logger used by Collector.
func WithLogFunc(c *Collector, logFunc LogFunc) *Collector { return c.WithLogFunc(logFunc) }

// WithCollectorLogFunc sets the logger during Collector construction.
func WithCollectorLogFunc(logFunc LogFunc) CollectorOption {
	return errimpl.WithCollectorLogFunc(logFunc)
}

// WithCollectorRunner sets the function used to launch Collector asynchronous work.
func WithCollectorRunner(runner func(func())) CollectorOption {
	return errimpl.WithCollectorRunner(runner)
}

// WithCollectorContext sets the context attached to log entries during Collector construction.
func WithCollectorContext(ctx context.Context) CollectorOption {
	return errimpl.WithCollectorContext(ctx)
}

// WithCollectorLevel sets the log level during Collector construction.
func WithCollectorLevel(level logrus.Level) CollectorOption {
	return errimpl.WithCollectorLevel(level)
}

// WithCollectorTimerFactory sets the timer factory during Collector construction.
func WithCollectorTimerFactory(factory TimerFactory) CollectorOption {
	return errimpl.WithCollectorTimerFactory(factory)
}

// WithCollectorStackCaptureOptions sets stack capture options during Collector construction.
func WithCollectorStackCaptureOptions(opts ...StackOption) CollectorOption {
	return errimpl.WithCollectorStackCaptureOptions(opts...)
}

// WithCollectorStackOptions sets stack capture options used by Collector logging.
func WithCollectorStackOptions(c *Collector, opts ...StackOption) *Collector {
	return c.WithStackOptions(opts...)
}

// WithWaitContext sets a context that can cancel a single WaitUntilWithOptions call.
func WithWaitContext(ctx context.Context) WaitOption { return errimpl.WithWaitContext(ctx) }

// WithWaitTimerFactory sets the timer factory for a single WaitUntilWithOptions call.
func WithWaitTimerFactory(factory TimerFactory) WaitOption {
	return errimpl.WithWaitTimerFactory(factory)
}

// WaitUntilWithOptions waits using per-call wait options.
func WaitUntilWithOptions(c *Collector, duration time.Duration, opts ...WaitOption) (bool, error) {
	return c.WaitUntilWithOptions(duration, opts...)
}
