package errx

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
)

func TestWithCollectorContext(t *testing.T) {
	silenceLogrus(t)

	ctx := context.WithValue(context.Background(), collectorContextKey{}, "value")
	c := NewCollectorWithOptions(WithCollectorContext(ctx))
	if c == nil {
		t.Fatal("NewCollectorWithOptions returned nil")
	}
}

func TestWithCollectorLevel(t *testing.T) {
	silenceLogrus(t)

	c := NewCollectorWithOptions(WithCollectorLevel(logrus.WarnLevel))
	if c == nil {
		t.Fatal("NewCollectorWithOptions with level returned nil")
	}
}

func TestWithCollectorTimerFactory(t *testing.T) {
	silenceLogrus(t)

	factory := func(d time.Duration) (<-chan time.Time, Timer) {
		ch := make(chan time.Time)
		return ch, &noopTimer{}
	}
	c := NewCollectorWithOptions(WithCollectorTimerFactory(factory))
	if c == nil {
		t.Fatal("NewCollectorWithOptions with timer factory returned nil")
	}
}

func TestWithCollectorLogFunc(t *testing.T) {
	silenceLogrus(t)

	var called bool
	c := NewCollectorWithOptions(WithCollectorLogFunc(func(ctx context.Context, level logrus.Level, err error, stack string, format string, args ...any) {
		called = true
	}))
	if c == nil {
		t.Fatal("NewCollectorWithOptions with log func returned nil")
	}
	// Trigger a log call
	_ = c.Recover(func() error { return errors.New("test") }, "test")
	if !called {
		t.Fatal("custom log func was not called")
	}
}

func TestNilCollectorOptionsDoNotClearPreviousProviders(t *testing.T) {
	silenceLogrus(t)

	timerFactory := func(time.Duration) (<-chan time.Time, Timer) {
		ch := make(chan time.Time)
		return ch, &noopTimer{}
	}
	var logged bool
	var runnerCalls int
	c := NewCollectorWithOptions(
		WithCollectorTimerFactory(timerFactory),
		WithCollectorTimerFactory(nil),
		WithCollectorLogFunc(func(context.Context, logrus.Level, error, string, string, ...any) {
			logged = true
		}),
		WithCollectorLogFunc(nil),
		WithCollectorRunner(func(fn func()) {
			runnerCalls++
			fn()
		}),
		WithCollectorRunner(nil),
	)

	if cfg := c.waitConfig(); cfg.timerFactory == nil {
		t.Fatal("nil WithCollectorTimerFactory cleared previous timer factory")
	}
	c.GoRun(func() error { return errors.New("boom") }, "job")
	if err := c.Error(); err == nil {
		t.Fatal("Collector.Error() = nil, want collected error")
	}
	if runnerCalls != 1 {
		t.Fatalf("nil WithCollectorRunner cleared previous runner: calls=%d", runnerCalls)
	}
	if !logged {
		t.Fatal("nil WithCollectorLogFunc cleared previous log func")
	}
}

func TestWithCollectorStackCaptureOptions(t *testing.T) {
	silenceLogrus(t)

	opts := []StackOption{WithDebugStackFunc(func() []byte {
		return []byte("custom stack")
	})}
	c := NewCollectorWithOptions(WithCollectorStackCaptureOptions(opts...))
	if c == nil {
		t.Fatal("NewCollectorWithOptions with stack options returned nil")
	}
}

func TestWithWaitTimerFactory(t *testing.T) {
	silenceLogrus(t)

	factory := func(d time.Duration) (<-chan time.Time, Timer) {
		ch := make(chan time.Time)
		close(ch) // close immediately so select picks it
		return ch, &noopTimer{}
	}
	var done bool
	done, _ = NewCollector().WaitUntilWithOptions(time.Second, WithWaitTimerFactory(factory))
	if done {
		t.Fatal("WaitUntilWithOptions with closed timer should return false")
	}
}

func TestNilWaitTimerFactoryDoesNotOverwriteConfiguredProvider(t *testing.T) {
	factory := func(time.Duration) (<-chan time.Time, Timer) {
		ch := make(chan time.Time)
		return ch, &noopTimer{}
	}
	cfg := NewCollector().waitConfig(WithWaitTimerFactory(factory), WithWaitTimerFactory(nil))
	if cfg.timerFactory == nil {
		t.Fatal("nil WithWaitTimerFactory should not overwrite configured timer factory")
	}
}

func TestWithWaitContext(t *testing.T) {
	silenceLogrus(t)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // already cancelled

	factory := func(d time.Duration) (<-chan time.Time, Timer) {
		ch := make(chan time.Time)
		return ch, &noopTimer{}
	}
	var done bool
	done, _ = NewCollector().WaitUntilWithOptions(time.Hour, WithWaitContext(ctx), WithWaitTimerFactory(factory))
	if done {
		t.Fatal("WaitUntilWithOptions with cancelled context should return false")
	}
}

func TestCollectorWithLogFunc(t *testing.T) {
	silenceLogrus(t)

	var called bool
	c := NewCollector().WithLogFunc(func(ctx context.Context, level logrus.Level, err error, stack string, format string, args ...any) {
		called = true
	})
	if c == nil {
		t.Fatal("WithLogFunc returned nil")
	}
	_ = c.Recover(func() error { return errors.New("test") }, "test")
	if !called {
		t.Fatal("custom log func was not called after WithLogFunc")
	}
}

func TestCollectorWithStackOptions(t *testing.T) {
	silenceLogrus(t)

	c := NewCollector().WithStackOptions(WithDebugStackFunc(func() []byte {
		return []byte("custom stack")
	}))
	if c == nil {
		t.Fatal("WithStackOptions returned nil")
	}
}

func TestWithDebugStackFunc(t *testing.T) {
	const want = "debug stack content"
	cfg := applyStackOptions([]StackOption{
		WithDebugStackFunc(func() []byte {
			return []byte(want)
		}),
	})
	if got := string(cfg.debugStack()); got != want {
		t.Fatalf("debugStack() = %q, want %q", got, want)
	}

	// nil option should be ignored
	cfg2 := applyStackOptions([]StackOption{WithDebugStackFunc(nil)})
	if cfg2.debugStack == nil {
		t.Fatal("nil WithDebugStackFunc should not overwrite the default")
	}

	// default
	cfg3 := applyStackOptions(nil)
	if cfg3.debugStack == nil {
		t.Fatal("default debugStack should not be nil")
	}
}

type noopTimer struct{}

func (t *noopTimer) Stop() bool { return true }
