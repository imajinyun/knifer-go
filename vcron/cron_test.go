package vcron_test

import (
	"bytes"
	"errors"
	"testing"
	"time"

	"github.com/imajinyun/go-knifer/vcron"
)

func TestFacadePatternParse(t *testing.T) {
	p, err := vcron.NewPattern("0 0 * * *")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p == nil {
		t.Fatal("expected non-nil pattern")
	}
}

func TestFacadePatternParseInvalid(t *testing.T) {
	_, err := vcron.NewPattern("invalid")
	if err == nil {
		t.Fatal("expected error for invalid pattern")
	}
}

func TestFacadePartConstants(t *testing.T) {
	if err := vcron.PartMinute.CheckValue(59); err != nil {
		t.Fatalf("PartMinute.CheckValue(59) = %v", err)
	}
	if err := vcron.PartMinute.CheckValue(60); err == nil {
		t.Fatal("PartMinute.CheckValue(60) should fail")
	}
	if !vcron.AlwaysTrueMatcher.Match(123) || vcron.AlwaysTrueMatcher.NextAfter(7) != 7 {
		t.Fatal("AlwaysTrueMatcher facade mismatch")
	}
}

func TestFacadeMustPattern(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for invalid pattern")
		}
	}()
	vcron.MustNewPattern("bad")
}

func TestFacadeSchedulerLifecycle(t *testing.T) {
	s := vcron.NewScheduler()
	if s == nil {
		t.Fatal("expected non-nil scheduler")
	}

	id, err := vcron.CronScheduleFunc("* * * * * *", func() {})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id == "" {
		t.Fatal("expected non-empty task id")
	}

	if !vcron.CronRemove(id) {
		t.Fatal("expected task to be removed")
	}
}

func TestFacadeSchedulerWithOptions(t *testing.T) {
	loc := time.FixedZone("facade", 8*60*60)
	now := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	s := vcron.NewSchedulerWithOptions(
		vcron.WithLocation(loc),
		vcron.WithMatchSecond(true),
		vcron.WithIDGenerator(func() string { return "facade-task" }),
		vcron.WithClock(func() time.Time { return now }),
		vcron.WithSleeper(func(d time.Duration, stopCh <-chan struct{}) bool {
			now = now.Add(d)
			return true
		}),
		vcron.WithExecutor(func(fn func()) { fn() }),
	)
	if s.Config().Location != loc {
		t.Fatalf("scheduler location = %v, want %v", s.Config().Location, loc)
	}
	if !s.IsMatchSecond() {
		t.Fatal("scheduler should match seconds")
	}
	id, err := s.ScheduleFunc("* * * * * *", func() {})
	if err != nil {
		t.Fatalf("ScheduleFunc with options: %v", err)
	}
	if id != "facade-task" {
		t.Fatalf("scheduled id = %q, want facade-task", id)
	}
}

func TestFacadeSchedulerIDRandomReaderOption(t *testing.T) {
	s := vcron.NewSchedulerWithOptions(vcron.WithIDRandomReader(bytes.NewReader([]byte{8, 7, 6, 5, 4, 3, 2, 1})))
	id, err := s.ScheduleFunc("* * * * *", func() {})
	if err != nil {
		t.Fatalf("ScheduleFunc: %v", err)
	}
	if id != "0807060504030201" {
		t.Fatalf("id = %q, want 0807060504030201", id)
	}
}

func TestFacadeDefaultSchedulerOptions(t *testing.T) {
	global := vcron.ConfigureDefaultScheduler(vcron.WithIDGenerator(func() string { return "global-id" }))
	t.Cleanup(func() { vcron.ConfigureDefaultScheduler() })
	isolated := vcron.NewSchedulerWithOptions(vcron.WithIDGenerator(func() string { return "facade-isolated" }))

	id, err := vcron.CronScheduleFuncWithOptions("* * * * *", func() {}, vcron.WithDefaultScheduler(isolated))
	if err != nil {
		t.Fatalf("CronScheduleFuncWithOptions: %v", err)
	}
	if id != "facade-isolated" || isolated.Size() != 1 || global.Size() != 0 {
		t.Fatalf("default scheduler option not isolated: id=%q isolated=%d global=%d", id, isolated.Size(), global.Size())
	}
	if !vcron.CronRemoveWithOptions(id, vcron.WithDefaultScheduler(isolated)) {
		t.Fatal("CronRemoveWithOptions should remove isolated task")
	}
}

func TestFacadeSchedulerStartedConfigErrors(t *testing.T) {
	s := vcron.NewSchedulerWithOptions(vcron.WithMatchSecond(true))
	if err := s.Start(); err != nil {
		t.Fatalf("Start: %v", err)
	}
	defer s.Stop()

	if err := s.SetMatchSecondE(false); !errors.Is(err, vcron.ErrSchedulerStarted) {
		t.Fatalf("SetMatchSecondE while started = %v, want ErrSchedulerStarted", err)
	}
	if err := s.SetTimeZoneE(time.UTC); !errors.Is(err, vcron.ErrSchedulerStarted) {
		t.Fatalf("SetTimeZoneE while started = %v, want ErrSchedulerStarted", err)
	}
	if err := vcron.CronSetMatchSecondEWithOptions(false, vcron.WithDefaultScheduler(s)); !errors.Is(err, vcron.ErrSchedulerStarted) {
		t.Fatalf("CronSetMatchSecondEWithOptions while started = %v, want ErrSchedulerStarted", err)
	}
	if !s.IsMatchSecond() {
		t.Fatal("started scheduler config should not be mutated")
	}
}

func TestFacadeConfig(t *testing.T) {
	cfg := vcron.NewConfig()
	if cfg == nil {
		t.Fatal("expected non-nil config")
	}
	loc := time.FixedZone("facade-config", 8*3600)
	cfg = vcron.NewConfigWithOptions(vcron.WithConfigLocation(loc), vcron.WithConfigMatchSecond(true))
	if cfg.Location != loc || !cfg.MatchSecond {
		t.Fatalf("NewConfigWithOptions = %#v", cfg)
	}
}
