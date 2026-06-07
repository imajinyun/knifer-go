package errx

import (
	"context"
	"errors"
	"io"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
)

type collectorContextKey struct{}

func TestCollectorRecoverCollectsReturnedError(t *testing.T) {
	silenceLogrus(t)

	want := errors.New("boom")
	c := NewCollector().WithContext(context.TODO()).WithLevel(logrus.WarnLevel)
	got := c.Recover(func() error { return want }, "run %s", "job")
	if !ErrorIs(got, want) {
		t.Fatalf("Recover() error = %v, want wrapped %v", got, want)
	}
	if !ErrorIs(c.Error(), want) {
		t.Fatalf("Collector.Error() should include %v", want)
	}
}

func TestCollectorRecoverConvertsPanicToError(t *testing.T) {
	silenceLogrus(t)

	c := NewCollector()
	got := c.Recover(func() error {
		panic("panic-value")
	}, "panic job")
	if got == nil || !strings.Contains(got.Error(), "panic-value") {
		t.Fatalf("Recover() panic error = %v, want panic value", got)
	}
	if err := c.Error(); err == nil || !strings.Contains(err.Error(), "panic-value") {
		t.Fatalf("Collector.Error() = %v, want panic value", err)
	}
}

func TestCollectorGoRunAggregatesConcurrentErrors(t *testing.T) {
	silenceLogrus(t)

	errA := errors.New("a")
	errB := errors.New("b")
	c := NewCollector()
	c.GoRun(func() error { return errA }, "job a")
	c.GoRun(func() error { return errB }, "job b")

	got := c.Error()
	if !ErrorIs(got, errA) || !ErrorIs(got, errB) {
		t.Fatalf("Collector.Error() = %v, want both errors", got)
	}
}

func TestCollectorGoRunUsesRunner(t *testing.T) {
	silenceLogrus(t)

	runnerCalls := 0
	c := NewCollectorWithOptions(WithCollectorRunner(func(fn func()) {
		runnerCalls++
		fn()
	}))
	c.GoRun(func() error { return nil }, "sync")
	if runnerCalls != 1 {
		t.Fatalf("runner calls = %d, want 1", runnerCalls)
	}
	if err := c.Error(); err != nil {
		t.Fatalf("Collector.Error() = %v, want nil", err)
	}
}

func TestCollectorWaitUntil(t *testing.T) {
	silenceLogrus(t)

	c := NewCollector()
	started := make(chan struct{})
	release := make(chan struct{})
	c.GoRun(func() error {
		close(started)
		<-release
		return nil
	}, "blocked")
	<-started

	done, err := c.WaitUntil(10 * time.Millisecond)
	if done || err != nil {
		t.Fatalf("WaitUntil() before release = (%v, %v), want (false, nil)", done, err)
	}
	close(release)
	done, err = c.WaitUntil(time.Second)
	if !done || err != nil {
		t.Fatalf("WaitUntil() after release = (%v, %v), want (true, nil)", done, err)
	}
}

type collectorTestTimer struct {
	stopped atomic.Bool
}

func (t *collectorTestTimer) Stop() bool {
	t.stopped.Store(true)
	return true
}

func TestCollectorWaitUntilWithTimerFactory(t *testing.T) {
	silenceLogrus(t)

	c := NewCollector()
	timerC := make(chan time.Time)
	timer := &collectorTestTimer{}
	c.WithTimerFactory(func(duration time.Duration) (<-chan time.Time, Timer) {
		if duration != time.Hour {
			t.Fatalf("timer duration = %s, want 1h", duration)
		}
		return timerC, timer
	})
	started := make(chan struct{})
	release := make(chan struct{})
	c.GoRun(func() error {
		close(started)
		<-release
		return nil
	}, "blocked")
	<-started

	doneC := make(chan bool, 1)
	go func() {
		done, err := c.WaitUntil(time.Hour)
		if err != nil {
			t.Errorf("WaitUntil() error = %v", err)
		}
		doneC <- done
	}()
	timerC <- time.Unix(1, 0)
	if done := <-doneC; done {
		t.Fatal("WaitUntil() done = true, want timeout")
	}
	if !timer.stopped.Load() {
		t.Fatal("custom timer was not stopped")
	}
	close(release)
}

func TestCollectorWaitUntilWithOptionsContext(t *testing.T) {
	silenceLogrus(t)

	c := NewCollector()
	started := make(chan struct{})
	release := make(chan struct{})
	c.GoRun(func() error {
		close(started)
		<-release
		return nil
	}, "blocked")
	<-started
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	done, err := c.WaitUntilWithOptions(time.Hour, WithWaitContext(ctx))
	if done || err != nil {
		t.Fatalf("WaitUntilWithOptions() = (%v, %v), want (false, nil)", done, err)
	}
	close(release)
}

func TestCollectorCollectErrorAliasAndNilFunctions(t *testing.T) {
	silenceLogrus(t)

	var called atomic.Bool
	c := NewCollector()
	c.CollectError(func() error {
		called.Store(true)
		return nil
	}, "alias")
	if !called.Load() {
		t.Fatal("CollectError() did not run the function")
	}
	if err := c.Recover(nil, "nil function"); err != nil {
		t.Fatalf("Recover(nil) error = %v, want nil", err)
	}
	if err := c.Error(); err != nil {
		t.Fatalf("Collector.Error() = %v, want nil", err)
	}
}

func silenceLogrus(t *testing.T) {
	t.Helper()
	logger := logrus.StandardLogger()
	oldOut := logger.Out
	oldFormatter := logger.Formatter
	oldLevel := logger.Level
	logger.SetOutput(io.Discard)
	logger.SetFormatter(EmptyFormatter)
	logger.SetLevel(logrus.TraceLevel)
	t.Cleanup(func() {
		logger.SetOutput(oldOut)
		logger.SetFormatter(oldFormatter)
		logger.SetLevel(oldLevel)
	})
}

func TestCollectorWithContextUsesProvidedContext(t *testing.T) {
	c := NewCollector()
	ctx := context.WithValue(context.Background(), collectorContextKey{}, "value")
	if got := c.WithContext(ctx); got != c {
		t.Fatal("WithContext should return the receiver")
	}
}
