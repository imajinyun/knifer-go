package errx

import (
	"context"
	"errors"
	"fmt"
	"runtime/debug"
	"sync"

	"github.com/hashicorp/go-multierror"
	knifer "github.com/imajinyun/knifer-go"
	"github.com/sirupsen/logrus"
)

// LogFunc writes an error log entry for errx helpers.
type LogFunc func(ctx context.Context, level logrus.Level, err error, stack string, format string, args ...any)

// DebugStackFunc captures the current goroutine stack.
type DebugStackFunc func() []byte

var defaultLogFuncState = struct {
	sync.RWMutex
	logFunc LogFunc
}{logFunc: logrusLogFunc}

type stackConfig struct {
	debugStack DebugStackFunc
}

// StackOption customizes fallback stack capture.
type StackOption func(*stackConfig)

// WithDebugStackFunc sets the function used when err does not carry a stack.
func WithDebugStackFunc(fn DebugStackFunc) StackOption {
	return func(c *stackConfig) {
		if fn != nil {
			c.debugStack = fn
		}
	}
}

func applyStackOptions(opts []StackOption) stackConfig {
	cfg := stackConfig{debugStack: debug.Stack}
	for _, opt := range opts {
		if opt != nil {
			opt(&cfg)
		}
	}
	if cfg.debugStack == nil {
		cfg.debugStack = debug.Stack
	}
	return cfg
}

// ConfigureDefaultLogFunc sets the package-level logger used when callers do
// not provide an explicit LogFunc. Passing nil restores the logrus-backed
// default logger.
func ConfigureDefaultLogFunc(logFunc LogFunc) {
	defaultLogFuncState.Lock()
	defer defaultLogFuncState.Unlock()
	if logFunc == nil {
		defaultLogFuncState.logFunc = logrusLogFunc
		return
	}
	defaultLogFuncState.logFunc = logFunc
}

// ResetDefaultLogFunc restores the package-level logger to the logrus-backed default.
func ResetDefaultLogFunc() { ConfigureDefaultLogFunc(nil) }

func getDefaultLogFunc() LogFunc {
	defaultLogFuncState.RLock()
	defer defaultLogFuncState.RUnlock()
	if defaultLogFuncState.logFunc == nil {
		return logrusLogFunc
	}
	return defaultLogFuncState.logFunc
}

func logrusLogFunc(ctx context.Context, level logrus.Level, err error, stack string, format string, args ...any) {
	logrus.WithContext(ctx).
		WithError(err).
		WithField("stack", stack).
		Logf(level, format, args...)
}

// WithStack is implemented by errors that can expose a string stack trace.
type WithStack interface {
	Stack() string
}

// PanicError wraps a recovered panic value and records the stack captured at the
// recovery point. If the panic value is an error, Unwrap exposes it for errors.Is
// and errors.As.
type PanicError struct {
	Value      any
	Cause      error
	StackTrace StackTrace
}

// ErrorCode returns the knifer-go error code for a recovered panic.
func (e *PanicError) ErrorCode() knifer.ErrCode {
	if e == nil {
		return ""
	}
	if code, ok := knifer.CodeOf(e.Cause); ok {
		return code
	}
	return knifer.ErrCodeInternal
}

// Error returns the recovered panic value as an error message.
func (e *PanicError) Error() string {
	if e == nil {
		return "<nil>"
	}
	if e.Cause != nil {
		return e.Cause.Error()
	}
	return fmt.Sprint(e.Value)
}

// Unwrap returns the panic value when it is an error.
func (e *PanicError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Cause
}

// Is supports errors.Is(err, knifer.ErrCodeXxx) matching by error code.
func (e *PanicError) Is(target error) bool {
	if e == nil || target == nil {
		return false
	}
	code, ok := target.(knifer.ErrCode)
	return ok && e.ErrorCode() == code
}

// Stack returns the stack captured when the panic was recovered.
func (e *PanicError) Stack() string {
	if e == nil || len(e.StackTrace) == 0 {
		return ""
	}
	return fmt.Sprintf("%+v", e.StackTrace)
}

// GetStack returns the stack attached to err, or the current goroutine stack.
func GetStack(err error) string {
	return GetStackWithOptions(err)
}

// GetStackWithOptions returns the stack attached to err, or captures a stack with options.
func GetStackWithOptions(err error, opts ...StackOption) string {
	if err == nil {
		return ""
	}
	var ws WithStack
	if errors.As(err, &ws) {
		return ws.Stack()
	}
	cfg := applyStackOptions(opts)
	return string(cfg.debugStack())
}

// ErrorIs is like errors.Is, but it also checks each member of a multierror.
func ErrorIs(err error, target error) bool {
	if target == nil {
		return err == nil
	}
	if errors.Is(err, target) {
		return true
	}
	var merr *multierror.Error
	if !errors.As(err, &merr) {
		return false
	}
	for _, item := range merr.Errors {
		if ErrorIs(item, target) {
			return true
		}
	}
	return false
}
