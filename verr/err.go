package verr

import (
	"context"
	"io"

	"github.com/sirupsen/logrus"

	errimpl "github.com/imajinyun/go-knifer/internal/errx"
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

// InitOption customizes logrus/Sentry initialization.
type InitOption = errimpl.InitOption

// Collector runs functions, recovers panics, logs failures, and aggregates errors.
type Collector = errimpl.Collector

// Wrapper executes a function with panic recovery and optional logging.
type Wrapper = errimpl.Wrapper

// NewCollector creates a Collector that logs failures at error level.
func NewCollector() *Collector { return errimpl.NewCollector() }

// GetStack returns the stack attached to err, or the current goroutine stack.
func GetStack(err error) string { return errimpl.GetStack(err) }

// ErrorIs is like errors.Is, but it also checks each member of a multierror.
func ErrorIs(err error, target error) bool { return errimpl.ErrorIs(err, target) }

// MustExit logs err and panics when err is non-nil.
func MustExit(ctx context.Context, err error) { errimpl.MustExit(ctx, err) }

// Init configures logrus output and optional Sentry forwarding.
func Init(sentryDSN string) { errimpl.Init(sentryDSN) }

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

// InitWithOptions configures logrus output and optional Sentry forwarding with options.
func InitWithOptions(opts ...InitOption) { errimpl.InitWithOptions(opts...) }

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
func GetStackTrace(skip int) StackTrace { return errimpl.GetStackTrace(skip) }

// WithStackSkip sets how many caller frames to skip while capturing a stack trace.
func WithStackSkip(skip int) StackTraceOption { return errimpl.WithStackSkip(skip) }

// WithStackDepth limits the number of captured stack frames.
func WithStackDepth(depth int) StackTraceOption { return errimpl.WithStackDepth(depth) }

// GetStackTraceWithOptions captures the current goroutine stack trace with options.
func GetStackTraceWithOptions(opts ...StackTraceOption) StackTrace {
	return errimpl.GetStackTraceWithOptions(opts...)
}

// WithLevel sets the log level used for recovered or returned errors.
func WithLevel(c *Collector, level logrus.Level) *Collector { return c.WithLevel(level) }
