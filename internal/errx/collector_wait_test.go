package errx

import (
	"context"
	"sync/atomic"
	"testing"
	"time"
)

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
