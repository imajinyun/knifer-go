package cron

import (
	"sync/atomic"
	"testing"
	"time"
)

func TestSchedulerOptions(t *testing.T) {
	loc := time.FixedZone("test", 8*3600)
	var submitted atomic.Int32
	var runnerCalls atomic.Int32
	var sleepCalls atomic.Int32
	now := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	s := NewSchedulerWithOptions(
		WithLocation(loc),
		WithMatchSecond(true),
		WithIDGenerator(func() string { return "custom-id" }),
		WithClock(func() time.Time { return now }),
		WithSleeper(func(d time.Duration, stopCh <-chan struct{}) bool {
			sleepCalls.Add(1)
			now = now.Add(d)
			select {
			case <-stopCh:
				return false
			default:
				return true
			}
		}),
		WithExecutor(func(fn func()) {
			submitted.Add(1)
			fn()
		}),
		WithRunner(func(fn func()) {
			runnerCalls.Add(1)
			go fn()
		}),
	)
	if s.Config().Location != loc || !s.IsMatchSecond() {
		t.Fatalf("scheduler options not applied: %#v", s.Config())
	}
	id, err := s.ScheduleFunc("* * * * * *", func() {})
	if err != nil {
		t.Fatalf("schedule with custom id: %v", err)
	}
	if id != "custom-id" {
		t.Fatalf("custom id = %q", id)
	}
	s.submit(func() {})
	if submitted.Load() != 1 {
		t.Fatalf("custom executor not used")
	}
	if s.nowMillis() != now.UnixMilli() {
		t.Fatalf("custom clock not used")
	}
	if !s.sleep(time.Millisecond, make(chan struct{})) || sleepCalls.Load() != 1 {
		t.Fatalf("custom sleeper not used")
	}
	if err := s.Start(); err != nil {
		t.Fatalf("start with custom runner: %v", err)
	}
	if runnerCalls.Load() != 1 {
		t.Fatalf("custom runner calls = %d, want 1", runnerCalls.Load())
	}
	s.Stop()
}
