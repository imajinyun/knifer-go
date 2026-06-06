package errx

import (
	"context"

	"github.com/sirupsen/logrus"
)

type exitConfig struct {
	logFunc   LogFunc
	panicFunc func(error)
}

// ExitOption customizes MustExitWithOptions.
type ExitOption func(*exitConfig)

// WithExitLogFunc sets the logger used by MustExitWithOptions.
func WithExitLogFunc(logFunc LogFunc) ExitOption {
	return func(c *exitConfig) {
		if logFunc != nil {
			c.logFunc = logFunc
		}
	}
}

// WithExitPanicFunc sets the panic function used by MustExitWithOptions.
func WithExitPanicFunc(panicFunc func(error)) ExitOption {
	return func(c *exitConfig) {
		if panicFunc != nil {
			c.panicFunc = panicFunc
		}
	}
}

func defaultPanicFunc(err error) { panic(err) }

func applyExitOptions(opts []ExitOption) exitConfig {
	cfg := exitConfig{logFunc: getDefaultLogFunc(), panicFunc: defaultPanicFunc}
	for _, opt := range opts {
		if opt != nil {
			opt(&cfg)
		}
	}
	if cfg.logFunc == nil {
		cfg.logFunc = getDefaultLogFunc()
	}
	if cfg.panicFunc == nil {
		cfg.panicFunc = defaultPanicFunc
	}
	return cfg
}

// MustExit logs err and panics when err is non-nil.
func MustExit(ctx context.Context, err error) {
	MustExitWithOptions(ctx, err)
}

// MustExitWithOptions logs err and panics when err is non-nil with custom options.
func MustExitWithOptions(ctx context.Context, err error, opts ...ExitOption) {
	if err == nil {
		return
	}
	cfg := applyExitOptions(opts)
	cfg.logFunc(ctx, logrus.ErrorLevel, err, GetStack(err), "exit with error")
	cfg.panicFunc(err)
}
