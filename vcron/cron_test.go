package vcron_test

import (
	"testing"
	"time"

	"github.com/imajinyun/go-knifer/vcron"
)

func TestFacadePatternParse(t *testing.T) {
	p, err := vcron.NewCronPattern("0 0 * * *")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p == nil {
		t.Fatal("expected non-nil pattern")
	}
}

func TestFacadePatternParseInvalid(t *testing.T) {
	_, err := vcron.NewCronPattern("invalid")
	if err == nil {
		t.Fatal("expected error for invalid pattern")
	}
}

func TestFacadeMustPattern(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for invalid pattern")
		}
	}()
	vcron.MustNewCronPattern("bad")
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
	s := vcron.NewSchedulerWithOptions(
		vcron.WithLocation(loc),
		vcron.WithMatchSecond(true),
		vcron.WithIDGenerator(func() string { return "facade-task" }),
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

func TestFacadeConfig(t *testing.T) {
	cfg := vcron.NewCronConfig()
	if cfg == nil {
		t.Fatal("expected non-nil config")
	}
}
