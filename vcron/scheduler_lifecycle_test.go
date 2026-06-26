package vcron_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/imajinyun/knifer-go/vcron"
)

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

func TestFacadeSchedulerControlledLifecycle(t *testing.T) {
	runs := 0
	stopped := false
	s := vcron.NewSchedulerWithOptions(
		vcron.WithMatchSecond(true),
		vcron.WithRunner(func(fn func()) {
			runs++
			fn()
		}),
		vcron.WithSleeper(func(time.Duration, <-chan struct{}) bool {
			stopped = true
			return false
		}),
	)

	if err := vcron.CronStartWithOptions(vcron.WithDefaultScheduler(s)); err != nil {
		t.Fatalf("CronStartWithOptions: %v", err)
	}
	if !s.IsStarted() || runs != 1 {
		t.Fatalf("scheduler should be started once, started=%v runs=%d", s.IsStarted(), runs)
	}
	if err := vcron.CronSetMatchSecondEWithOptions(false, vcron.WithDefaultScheduler(s)); !errors.Is(err, vcron.ErrSchedulerStarted) {
		t.Fatalf("CronSetMatchSecondEWithOptions while started = %v", err)
	}
	if err := vcron.CronStartWithOptions(vcron.WithDefaultScheduler(s)); err == nil {
		t.Fatal("second CronStartWithOptions should fail")
	}
	vcron.CronStopWithOptions(vcron.WithDefaultScheduler(s))
	if s.IsStarted() || !stopped {
		t.Fatalf("scheduler should be stopped, started=%v stopped=%v", s.IsStarted(), stopped)
	}
	if err := vcron.CronSetMatchSecondEWithOptions(false, vcron.WithDefaultScheduler(s)); err != nil || s.IsMatchSecond() {
		t.Fatalf("CronSetMatchSecondEWithOptions after stop err=%v matchSecond=%v", err, s.IsMatchSecond())
	}
}

func TestFacadeSchedulerShutdownRestartAndCounts(t *testing.T) {
	runs := 0
	s := vcron.NewSchedulerWithOptions(
		vcron.WithRunner(func(fn func()) {
			runs++
			go fn()
		}),
		vcron.WithSleeper(func(time.Duration, <-chan struct{}) bool { return false }),
	)

	if err := vcron.CronRestartWithOptions(vcron.WithDefaultScheduler(s)); err != nil {
		t.Fatalf("CronRestartWithOptions: %v", err)
	}
	if !s.IsStarted() || runs != 1 {
		t.Fatalf("restart should start scheduler, started=%v runs=%d", s.IsStarted(), runs)
	}
	if vcron.CronRunningCount() < 0 || vcron.CronLaunchingCount() < 0 {
		t.Fatal("default running/launching counts should be non-negative")
	}
	if err := vcron.CronShutdownWithOptions(context.Background(), vcron.WithDefaultScheduler(s)); err != nil {
		t.Fatalf("CronShutdownWithOptions: %v", err)
	}
	if s.IsStarted() {
		t.Fatal("shutdown should stop scheduler")
	}
}

func TestFacadeDefaultSchedulerLifecycle(t *testing.T) {
	vcron.ConfigureDefaultScheduler(
		vcron.WithRunner(func(fn func()) { go fn() }),
		vcron.WithSleeper(func(time.Duration, <-chan struct{}) bool { return false }),
	)
	t.Cleanup(func() { vcron.ConfigureDefaultScheduler() })

	vcron.CronSetMatchSecond(true)
	if err := vcron.CronSetMatchSecondE(true); err != nil {
		t.Fatalf("CronSetMatchSecondE before start: %v", err)
	}
	if err := vcron.CronStart(); err != nil {
		t.Fatalf("CronStart: %v", err)
	}
	vcron.CronStop()
	if err := vcron.CronRestart(); err != nil {
		t.Fatalf("CronRestart: %v", err)
	}
	if err := vcron.CronShutdown(context.Background(), true); err != nil {
		t.Fatalf("CronShutdown: %v", err)
	}
}
