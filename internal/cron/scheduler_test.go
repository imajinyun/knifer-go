package cron

import (
	"errors"
	"sync/atomic"
	"testing"
	"time"

	knifer "github.com/imajinyun/go-knifer"
)

func TestSchedulerLifecycle(t *testing.T) {
	s := NewScheduler()
	s.SetMatchSecond(true)

	var counter atomic.Int32
	id, err := s.ScheduleFunc("* * * * * *", func() { counter.Add(1) })
	if err != nil {
		t.Fatalf("schedule: %v", err)
	}
	if id == "" {
		t.Fatalf("empty id")
	}
	if s.Size() != 1 {
		t.Fatalf("expect 1 task, got %d", s.Size())
	}
	if err := s.Start(); err != nil {
		t.Fatalf("start: %v", err)
	}
	defer s.Stop()
	// Wait for at least two executions.
	time.Sleep(2500 * time.Millisecond)
	if counter.Load() < 1 {
		t.Fatalf("expect counter >= 1, got %d", counter.Load())
	}
	if !s.Deschedule(id) {
		t.Fatalf("expect deschedule ok")
	}
	if s.Size() != 0 {
		t.Fatalf("expect empty after deschedule")
	}
}

func TestSchedulerUpdatePattern(t *testing.T) {
	s := NewScheduler()
	if err := s.ScheduleWithID("t1", "0 0 * * *", TaskFunc(func() {})); err != nil {
		t.Fatalf("schedule: %v", err)
	}
	if err := s.UpdatePattern("t1", "0 12 * * *"); err != nil {
		t.Fatalf("update: %v", err)
	}
	if got := s.GetPattern("t1").Raw(); got != "0 12 * * *" {
		t.Fatalf("expect updated pattern, got %q", got)
	}
	if err := s.UpdatePattern("nx", "* * * * *"); err == nil {
		t.Fatalf("expect error for unknown id")
	}
}

func TestSchedulerDuplicateID(t *testing.T) {
	s := NewScheduler()
	if err := s.ScheduleWithID("a", "* * * * *", TaskFunc(func() {})); err != nil {
		t.Fatalf("schedule: %v", err)
	}
	if err := s.ScheduleWithID("a", "* * * * *", TaskFunc(func() {})); err == nil {
		t.Fatalf("expect duplicate id error")
	}
}

func TestSchedulerSchedulePatternRejectsNil(t *testing.T) {
	s := NewScheduler()
	err := s.SchedulePattern("nil", nil, TaskFunc(func() {}))
	if err == nil {
		t.Fatal("SchedulePattern nil pattern error = nil")
	}
	if !errors.Is(err, knifer.ErrCodeInvalidInput) {
		t.Fatalf("errors.Is(%v, %s) = false", err, knifer.ErrCodeInvalidInput)
	}
}

func TestSchedulerStartTwice(t *testing.T) {
	s := NewScheduler()
	if err := s.Start(); err != nil {
		t.Fatalf("start: %v", err)
	}
	defer s.Stop()
	if err := s.Start(); err == nil {
		t.Fatalf("expect error on second start")
	}
}

func TestSchedulerConfigSettersIgnoredWhileStarted(t *testing.T) {
	loc := time.FixedZone("before", 3600)
	after := time.FixedZone("after", 7200)
	s := NewSchedulerWithOptions(WithLocation(loc), WithMatchSecond(true))
	if err := s.Start(); err != nil {
		t.Fatalf("start: %v", err)
	}
	if err := s.SetMatchSecondE(false); !errors.Is(err, ErrSchedulerStarted) {
		t.Fatalf("SetMatchSecondE while started = %v, want ErrSchedulerStarted", err)
	} else if !errors.Is(err, knifer.ErrCodeUnsupported) {
		t.Fatalf("SetMatchSecondE while started = %v, want ErrCodeUnsupported", err)
	}
	if err := s.SetTimeZoneE(after); !errors.Is(err, ErrSchedulerStarted) {
		t.Fatalf("SetTimeZoneE while started = %v, want ErrSchedulerStarted", err)
	}
	s.SetMatchSecond(false).SetTimeZone(after)
	cfg := s.Config()
	if cfg.Location != loc || !cfg.MatchSecond {
		t.Fatalf("started scheduler config mutated: %#v", cfg)
	}
	s.Stop()
	if err := s.SetMatchSecondE(false); err != nil {
		t.Fatalf("SetMatchSecondE after stop: %v", err)
	}
	if err := s.SetTimeZoneE(after); err != nil {
		t.Fatalf("SetTimeZoneE after stop: %v", err)
	}
	s.SetMatchSecond(false).SetTimeZone(after)
	cfg = s.Config()
	if cfg.Location != after || cfg.MatchSecond {
		t.Fatalf("stopped scheduler config not updated: %#v", cfg)
	}
}

func TestSchedulerConfigReturnsSnapshot(t *testing.T) {
	s := NewSchedulerWithOptions(WithMatchSecond(true))
	cfg := s.Config()
	cfg.MatchSecond = false
	if !s.IsMatchSecond() {
		t.Fatal("mutating Config snapshot changed scheduler")
	}
}

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
