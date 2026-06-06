package errx

import (
	"context"

	"github.com/hashicorp/go-multierror"
	"github.com/sirupsen/logrus"
)

// EmptyFormatter suppresses logrus output while still allowing hooks to run.
var EmptyFormatter logrus.Formatter = &emptyFormatter{}

type emptyFormatter struct{}

// Format implements logrus.Formatter.
func (*emptyFormatter) Format(*logrus.Entry) ([]byte, error) { return []byte{}, nil }

// Wrapper executes a function with panic recovery and optional logging.
type Wrapper struct {
	f func() error

	log          bool
	level        logrus.Level
	format       string
	args         []any
	logger       LogFunc
	stackOptions []StackOption
}

// Wrap creates a recoverable function wrapper.
func Wrap(f func() error) *Wrapper { return &Wrapper{f: f} }

// WithInfof logs a recovered or returned error at info level.
func (w *Wrapper) WithInfof(format string, args ...any) *Wrapper {
	w.setLog(logrus.InfoLevel, format, args...)
	return w
}

// WithWarnf logs a recovered or returned error at warning level.
func (w *Wrapper) WithWarnf(format string, args ...any) *Wrapper {
	w.setLog(logrus.WarnLevel, format, args...)
	return w
}

// WithErrorf logs a recovered or returned error at error level.
func (w *Wrapper) WithErrorf(format string, args ...any) *Wrapper {
	w.setLog(logrus.ErrorLevel, format, args...)
	return w
}

// Exec executes the wrapped function and converts panics to errors.
func (w *Wrapper) Exec(ctx context.Context) (err error) {
	if ctx == nil {
		ctx = context.Background()
	}
	defer func() {
		if v := recover(); v != nil {
			err = multierror.Append(err, panicError(v))
		}
		if err != nil && w.log {
			logger := w.logger
			if logger == nil {
				logger = getDefaultLogFunc()
			}
			logger(ctx, w.level, err, GetStackWithOptions(err, w.stackOptions...), w.format, w.args...)
		}
	}()
	if w.f == nil {
		return nil
	}
	return w.f()
}

// WithLogFunc sets the logger used by this wrapper.
func (w *Wrapper) WithLogFunc(logFunc LogFunc) *Wrapper {
	if logFunc != nil {
		w.logger = logFunc
	}
	return w
}

// WithStackOptions sets stack capture options used by wrapper logging.
func (w *Wrapper) WithStackOptions(opts ...StackOption) *Wrapper {
	w.stackOptions = append([]StackOption(nil), opts...)
	return w
}

// Recover executes f with panic recovery and logs failures at error level.
func Recover(f func() error, format string, args ...any) error {
	return Wrap(f).WithErrorf(format, args...).Exec(context.Background())
}

// RecoverWithoutError executes f with panic recovery and logs failures at error level.
func RecoverWithoutError(f func(), format string, args ...any) error {
	if f == nil {
		return Wrap(nil).WithErrorf(format, args...).Exec(context.Background())
	}
	return Wrap(func() error { f(); return nil }).WithErrorf(format, args...).Exec(context.Background())
}

func (w *Wrapper) setLog(level logrus.Level, format string, args ...any) {
	w.log = true
	w.level = level
	w.format = format
	w.args = args
}
