// Package errx provides small error handling and panic-recovery helpers used by
// internal packages.
package errx

import (
	"context"
	"slices"
	"sync"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/sirupsen/logrus"
)

// Collector runs functions, recovers panics, logs failures, and aggregates
// returned errors. It is safe for concurrent use.
type Collector struct {
	level        logrus.Level
	ctx          context.Context
	timerFactory TimerFactory
	logFunc      LogFunc
	runner       func(func())
	stackOptions []StackOption

	swg sync.WaitGroup
	mux sync.Mutex
	err []error
}

// NewCollector creates a Collector that logs failures at error level.
func NewCollector() *Collector {
	return &Collector{
		level:        logrus.ErrorLevel,
		ctx:          context.Background(),
		timerFactory: newCollectorTimer,
		logFunc:      getDefaultLogFunc(),
		runner:       defaultCollectorRunner,
	}
}

// CollectorOption customizes Collector construction.
type CollectorOption func(*Collector)

// NewCollectorWithOptions creates a Collector customized by options.
func NewCollectorWithOptions(opts ...CollectorOption) *Collector {
	c := NewCollector()
	for _, opt := range opts {
		if opt != nil {
			opt(c)
		}
	}
	return c
}

func defaultCollectorRunner(fn func()) { go fn() }

// WithCollectorContext sets the context attached to log entries during Collector construction.
func WithCollectorContext(ctx context.Context) CollectorOption {
	return func(c *Collector) { c.WithContext(ctx) }
}

// WithCollectorLevel sets the log level used for recovered or returned errors during Collector construction.
func WithCollectorLevel(level logrus.Level) CollectorOption {
	return func(c *Collector) { c.WithLevel(level) }
}

// WithCollectorTimerFactory sets the default timer factory during Collector construction.
func WithCollectorTimerFactory(factory TimerFactory) CollectorOption {
	return func(c *Collector) { c.WithTimerFactory(factory) }
}

// WithCollectorLogFunc sets the logger during Collector construction.
func WithCollectorLogFunc(logFunc LogFunc) CollectorOption {
	return func(c *Collector) { c.WithLogFunc(logFunc) }
}

// WithCollectorRunner sets the function used to launch Collector asynchronous work.
func WithCollectorRunner(runner func(func())) CollectorOption {
	return func(c *Collector) { c.WithRunner(runner) }
}

// WithCollectorStackCaptureOptions sets stack capture options during Collector construction.
func WithCollectorStackCaptureOptions(opts ...StackOption) CollectorOption {
	return func(c *Collector) { c.WithStackOptions(opts...) }
}

// Timer stops a wait timer created by TimerFactory.
type Timer interface {
	Stop() bool
}

// TimerFactory creates a timer channel and stopper for Collector waits.
type TimerFactory func(time.Duration) (<-chan time.Time, Timer)

type waitConfig struct {
	ctx          context.Context
	timerFactory TimerFactory
}

// WaitOption customizes a single Collector wait call.
type WaitOption func(*waitConfig)

// WithWaitContext sets a context that can cancel a single WaitUntilWithOptions call.
func WithWaitContext(ctx context.Context) WaitOption {
	return func(c *waitConfig) { c.ctx = ctx }
}

// WithWaitTimerFactory sets the timer factory for a single WaitUntilWithOptions call.
func WithWaitTimerFactory(factory TimerFactory) WaitOption {
	return func(c *waitConfig) {
		if factory != nil {
			c.timerFactory = factory
		}
	}
}

// WithContext sets the context attached to log entries.
func (c *Collector) WithContext(ctx context.Context) *Collector {
	if ctx == nil {
		ctx = context.Background()
	}
	c.mux.Lock()
	defer c.mux.Unlock()
	c.ctx = ctx
	return c
}

// WithLevel sets the log level used for recovered or returned errors.
func (c *Collector) WithLevel(level logrus.Level) *Collector {
	c.mux.Lock()
	defer c.mux.Unlock()
	c.level = level
	return c
}

// WithTimerFactory sets the default timer factory used by WaitUntil.
func (c *Collector) WithTimerFactory(factory TimerFactory) *Collector {
	c.mux.Lock()
	defer c.mux.Unlock()
	if factory != nil {
		c.timerFactory = factory
	}
	return c
}

// WithLogFunc sets the logger used for recovered or returned errors.
func (c *Collector) WithLogFunc(logFunc LogFunc) *Collector {
	c.mux.Lock()
	defer c.mux.Unlock()
	if logFunc != nil {
		c.logFunc = logFunc
	}
	return c
}

// WithRunner sets the runner used by GoRun.
func (c *Collector) WithRunner(runner func(func())) *Collector {
	c.mux.Lock()
	defer c.mux.Unlock()
	if runner != nil {
		c.runner = runner
	}
	return c
}

// WithStackOptions sets stack capture options used by Collector logging.
func (c *Collector) WithStackOptions(opts ...StackOption) *Collector {
	c.mux.Lock()
	defer c.mux.Unlock()
	c.stackOptions = slices.Clone(opts)
	return c
}

// Collect stores err for the final aggregated result.
func (c *Collector) Collect(err error) {
	if err == nil {
		return
	}
	c.mux.Lock()
	c.err = append(c.err, err)
	c.mux.Unlock()
}

// Error waits for all launched functions and returns all collected errors.
func (c *Collector) Error() error {
	c.swg.Wait()
	return c.error()
}

// WaitUntil waits until all launched functions finish or duration expires.
// It returns whether all functions completed and the aggregated error, if any.
func (c *Collector) WaitUntil(duration time.Duration) (bool, error) {
	return c.WaitUntilWithOptions(duration)
}

// WaitUntilWithOptions waits until all launched functions finish, duration expires,
// or the optional wait context is canceled.
func (c *Collector) WaitUntilWithOptions(duration time.Duration, opts ...WaitOption) (bool, error) {
	if duration <= 0 {
		return false, nil
	}
	cfg := c.waitConfig(opts...)
	timerC, timer := cfg.timerFactory(duration)
	defer timer.Stop()

	select {
	case <-cfg.ctx.Done():
		return false, nil
	default:
	}
	done := make(chan struct{})
	go func() {
		c.swg.Wait()
		close(done)
	}()

	select {
	case <-done:
		return true, c.error()
	case <-timerC:
		return false, nil
	case <-cfg.ctx.Done():
		return false, nil
	}
}

func (c *Collector) waitConfig(opts ...WaitOption) waitConfig {
	c.mux.Lock()
	cfg := waitConfig{ctx: context.Background(), timerFactory: c.timerFactory}
	c.mux.Unlock()
	for _, opt := range opts {
		if opt != nil {
			opt(&cfg)
		}
	}
	if cfg.ctx == nil {
		cfg.ctx = context.Background()
	}
	if cfg.timerFactory == nil {
		cfg.timerFactory = newCollectorTimer
	}
	return cfg
}

func newCollectorTimer(duration time.Duration) (<-chan time.Time, Timer) {
	timer := time.NewTimer(duration)
	return timer.C, timer
}

func (c *Collector) currentRunner() func(func()) {
	c.mux.Lock()
	runner := c.runner
	c.mux.Unlock()
	if runner != nil {
		return runner
	}
	return defaultCollectorRunner
}

// Recover executes f in the current goroutine, recovers panics, logs failures,
// and stores non-nil errors in the collector.
func (c *Collector) Recover(f func() error, format string, args ...any) error {
	c.swg.Add(1)
	defer c.swg.Done()

	err := c.run(f, format, args...)
	c.Collect(err)
	return err
}

// GoRun executes f in a new goroutine and stores any panic or returned error.
func (c *Collector) GoRun(f func() error, format string, args ...any) {
	c.swg.Add(1)
	c.currentRunner()(func() {
		defer c.swg.Done()
		c.Collect(c.run(f, format, args...))
	})
}

// CollectError is kept as a compatibility alias for Recover.
func (c *Collector) CollectError(f func() error, format string, args ...any) {
	_ = c.Recover(f, format, args...)
}

func (c *Collector) run(f func() error, format string, args ...any) (err error) {
	defer func() {
		if v := recover(); v != nil {
			err = multierror.Append(err, panicError(v))
		}
		if err != nil {
			c.log(err, format, args...)
		}
	}()
	if f == nil {
		return nil
	}
	return f()
}

func (c *Collector) log(err error, format string, args ...any) {
	if format == "" {
		format = "operation failed"
	}
	c.mux.Lock()
	ctx, level, logFunc := c.ctx, c.level, c.logFunc
	stackOptions := slices.Clone(c.stackOptions)
	c.mux.Unlock()
	if logFunc == nil {
		logFunc = getDefaultLogFunc()
	}
	logFunc(ctx, level, err, GetStackWithOptions(err, stackOptions...), format, args...)
}

func (c *Collector) error() error {
	c.mux.Lock()
	defer c.mux.Unlock()
	if len(c.err) == 0 {
		return nil
	}
	return multierror.Append(nil, c.err...)
}

func panicError(v any) error {
	pe := &PanicError{
		Value:      v,
		StackTrace: GetStackTrace(4),
	}
	if err, ok := v.(error); ok {
		pe.Cause = err
	}
	return pe
}
