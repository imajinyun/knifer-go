package verr_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/imajinyun/knifer-go/verr"
)

func TestCollectorFacade(t *testing.T) {
	want := errors.New("collector failure")
	c := verr.NewCollector().WithContext(context.Background())
	c.GoRun(func() error { return want }, "collector")
	if got := c.Error(); !verr.ErrorIs(got, want) {
		t.Fatalf("Collector.Error() = %v, want %v", got, want)
	}
}

type facadeTimer struct{}

func (facadeTimer) Stop() bool { return true }

func TestCollectorWaitOptionsFacade(t *testing.T) {
	c := verr.NewCollector()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	called := false
	done, err := verr.WaitUntilWithOptions(c, time.Second,
		verr.WithWaitContext(ctx),
		verr.WithWaitTimerFactory(func(duration time.Duration) (<-chan time.Time, verr.Timer) {
			called = true
			if duration != time.Second {
				t.Fatalf("timer duration = %s, want 1s", duration)
			}
			return make(chan time.Time), facadeTimer{}
		}),
	)
	if done || err != nil {
		t.Fatalf("WaitUntilWithOptions() = (%v, %v), want (false, nil)", done, err)
	}
	if !called {
		t.Fatal("facade wait timer factory was not called")
	}
}

func TestCollectorOptionsFacadeUsesRunnerAndStackProvider(t *testing.T) {
	want := errors.New("collector option failure")
	ctx := context.WithValue(context.Background(), collectorContextKey{}, "facade")
	var ran bool
	var loggedErr error
	var loggedStack string
	var loggedLevel logrus.Level
	c := verr.NewCollectorWithOptions(
		verr.WithCollectorContext(ctx),
		verr.WithCollectorLevel(logrus.WarnLevel),
		verr.WithCollectorRunner(func(fn func()) {
			ran = true
			fn()
		}),
		verr.WithCollectorStackCaptureOptions(verr.WithDebugStackFunc(func() []byte { return []byte("collector facade stack") })),
		verr.WithCollectorLogFunc(func(gotCtx context.Context, level logrus.Level, err error, stack string, format string, args ...any) {
			if gotCtx.Value(collectorContextKey{}) != "facade" {
				t.Fatalf("collector context value = %v", gotCtx.Value(collectorContextKey{}))
			}
			loggedErr = err
			loggedStack = stack
			loggedLevel = level
		}),
	)
	c.GoRun(func() error { return want }, "collector facade")
	if !ran {
		t.Fatal("collector runner was not used")
	}
	if got := c.Error(); !verr.ErrorIs(got, want) {
		t.Fatalf("Collector.Error() = %v, want %v", got, want)
	}
	if !verr.ErrorIs(loggedErr, want) || loggedStack != "collector facade stack" || loggedLevel != logrus.WarnLevel {
		t.Fatalf("collector log err=%v stack=%q level=%s", loggedErr, loggedStack, loggedLevel)
	}
}

type collectorContextKey struct{}
