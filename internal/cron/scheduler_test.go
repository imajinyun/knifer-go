package cron

import (
	"bytes"
	"sync/atomic"
	"testing"
	"time"
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

func TestSchedulerIDRandomReaderOption(t *testing.T) {
	s := NewSchedulerWithOptions(WithIDRandomReader(bytes.NewReader([]byte{0, 1, 2, 3, 4, 5, 6, 7})))
	id, err := s.ScheduleFunc("* * * * *", func() {})
	if err != nil {
		t.Fatalf("ScheduleFunc: %v", err)
	}
	if id != "0001020304050607" {
		t.Fatalf("id = %q, want 0001020304050607", id)
	}
}

func TestSchedulerListener(t *testing.T) {
	s := NewScheduler()
	s.SetMatchSecond(true)

	var started, succ, failed atomic.Int32
	s.AddListener(&testListener{
		started: &started, succ: &succ, failed: &failed,
	})

	_, err := s.ScheduleFunc("* * * * * *", func() {})
	if err != nil {
		t.Fatalf("schedule: %v", err)
	}
	_, err = s.ScheduleFunc("* * * * * *", func() { panic("boom") })
	if err != nil {
		t.Fatalf("schedule: %v", err)
	}
	if err := s.Start(); err != nil {
		t.Fatalf("start: %v", err)
	}
	defer s.Stop()
	time.Sleep(1500 * time.Millisecond)
	if started.Load() < 2 {
		t.Fatalf("expect started >= 2, got %d", started.Load())
	}
	if succ.Load() < 1 {
		t.Fatalf("expect succ >= 1")
	}
	if failed.Load() < 1 {
		t.Fatalf("expect failed >= 1")
	}
}

type testListener struct {
	started *atomic.Int32
	succ    *atomic.Int32
	failed  *atomic.Int32
}

func (l *testListener) OnStart(*TaskExecutor)       { l.started.Add(1) }
func (l *testListener) OnSucceeded(*TaskExecutor)   { l.succ.Add(1) }
func (l *testListener) OnFailed(*TaskExecutor, any) { l.failed.Add(1) }

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

func TestDefaultSchedulerOperationOptions(t *testing.T) {
	global := ConfigureDefaultScheduler(WithIDGenerator(func() string { return "global-id" }))
	t.Cleanup(func() { ConfigureDefaultScheduler() })
	isolated := NewSchedulerWithOptions(WithIDGenerator(func() string { return "isolated-id" }))

	id, err := ScheduleFuncWithOptions("* * * * *", func() {}, WithDefaultScheduler(isolated))
	if err != nil {
		t.Fatalf("ScheduleFuncWithOptions: %v", err)
	}
	if id != "isolated-id" || isolated.Size() != 1 || global.Size() != 0 {
		t.Fatalf("default scheduler option not isolated: id=%q isolated=%d global=%d", id, isolated.Size(), global.Size())
	}
	if err := UpdatePatternWithOptions(id, "0 0 * * *", WithDefaultScheduler(isolated)); err != nil {
		t.Fatalf("UpdatePatternWithOptions: %v", err)
	}
	if !RemoveWithOptions(id, WithDefaultScheduler(isolated)) || isolated.Size() != 0 {
		t.Fatalf("RemoveWithOptions did not remove isolated task")
	}
}

func TestWithDefaultSchedulerOptionsCreatesPerCallScheduler(t *testing.T) {
	global := ConfigureDefaultScheduler(WithIDGenerator(func() string { return "global-id" }))
	t.Cleanup(func() { ConfigureDefaultScheduler() })

	id, err := ScheduleFuncWithOptions("* * * * *", func() {}, WithDefaultSchedulerOptions(WithIDGenerator(func() string { return "per-call-id" })))
	if err != nil {
		t.Fatalf("ScheduleFuncWithOptions: %v", err)
	}
	if id != "per-call-id" || global.Size() != 0 {
		t.Fatalf("per-call scheduler option leaked to global: id=%q globalSize=%d", id, global.Size())
	}
}

func TestConfigOptions(t *testing.T) {
	loc := time.FixedZone("config", 9*3600)
	cfg := NewConfigWithOptions(WithConfigLocation(loc), WithConfigMatchSecond(true))
	if cfg.Location != loc || !cfg.MatchSecond {
		t.Fatalf("config options not applied: %#v", cfg)
	}
	cfg = NewConfigWithOptions(WithConfigLocation(nil))
	if cfg.Location == nil {
		t.Fatal("nil config location should fall back to local")
	}
}
