package verr_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/imajinyun/go-knifer/verr"
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
