package verr

import (
	"context"

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

// WithLevel sets the log level used for recovered or returned errors.
func WithLevel(c *Collector, level logrus.Level) *Collector { return c.WithLevel(level) }
