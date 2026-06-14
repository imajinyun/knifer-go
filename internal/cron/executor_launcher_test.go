package cron

import (
	"context"
	"sync/atomic"
	"testing"
	"time"
)

func TestSchedulerLauncherPanicsAreIsolated(t *testing.T) {
	s := NewSchedulerWithOptions(WithExecutor(func(fn func()) { fn() }))
	if err := s.taskTable.Add("bad", nil, TaskFunc(func() {})); err != nil {
		t.Fatalf("add invalid task: %v", err)
	}
	s.launcherMgr.spawn(time.Now().UnixMilli())
	if got := s.LaunchingCount(); got != 0 {
		t.Fatalf("LaunchingCount = %d, want 0", got)
	}
}

func TestSchedulerShutdownWaitsForLaunchersBeforeExecutors(t *testing.T) {
	launcherStarted := make(chan struct{})
	allowLauncher := make(chan struct{})
	taskDone := make(chan struct{})
	var launcherSeen atomic.Bool

	s := NewSchedulerWithOptions(
		WithMatchSecond(true),
		WithClock(func() time.Time { return time.Unix(1, 0) }),
		WithSleeper(func(time.Duration, <-chan struct{}) bool { return true }),
		WithExecutor(func(fn func()) {
			if !launcherSeen.Swap(true) {
				close(launcherStarted)
				<-allowLauncher
			}
			go fn()
		}),
	)
	if _, err := s.ScheduleFunc("* * * * * *", func() { close(taskDone) }); err != nil {
		t.Fatalf("schedule: %v", err)
	}
	if err := s.Start(); err != nil {
		t.Fatalf("start: %v", err)
	}

	select {
	case <-launcherStarted:
	case <-time.After(time.Second):
		t.Fatal("launcher did not start")
	}
	if got := s.LaunchingCount(); got != 1 {
		t.Fatalf("LaunchingCount = %d, want 1", got)
	}

	shutdownDone := make(chan error, 1)
	go func() { shutdownDone <- s.Shutdown(context.Background()) }()
	select {
	case err := <-shutdownDone:
		t.Fatalf("Shutdown returned before launcher was released: %v", err)
	case <-time.After(20 * time.Millisecond):
	}

	close(allowLauncher)
	select {
	case err := <-shutdownDone:
		if err != nil {
			t.Fatalf("Shutdown error = %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("Shutdown did not finish after launcher was released")
	}
	select {
	case <-taskDone:
	case <-time.After(time.Second):
		t.Fatal("task did not run before Shutdown returned")
	}
	if got := s.LaunchingCount(); got != 0 {
		t.Fatalf("LaunchingCount after shutdown = %d, want 0", got)
	}
}
