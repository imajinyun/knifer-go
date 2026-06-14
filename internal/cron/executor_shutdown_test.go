package cron

import (
	"context"
	"testing"
	"time"
)

func TestSchedulerShutdownWaitsForRunningTasks(t *testing.T) {
	start := make(chan struct{})
	finish := make(chan struct{})
	s := NewSchedulerWithOptions(WithExecutor(func(fn func()) { go fn() }))
	s.executorMgr.spawn(NewCronTask("slow", MustNewPattern("* * * * *"), TaskFunc(func() {
		close(start)
		<-finish
	})))
	<-start
	if got := s.RunningCount(); got != 1 {
		t.Fatalf("RunningCount = %d, want 1", got)
	}
	done := make(chan error, 1)
	go func() { done <- s.Shutdown(context.Background()) }()
	select {
	case err := <-done:
		t.Fatalf("Shutdown returned before task finished: %v", err)
	case <-time.After(20 * time.Millisecond):
	}
	close(finish)
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("Shutdown error: %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("Shutdown did not return after task finished")
	}
	if got := s.RunningCount(); got != 0 {
		t.Fatalf("RunningCount after shutdown = %d, want 0", got)
	}
}

func TestSchedulerShutdownContextTimeout(t *testing.T) {
	start := make(chan struct{})
	finish := make(chan struct{})
	s := NewSchedulerWithOptions(WithExecutor(func(fn func()) { go fn() }))
	s.executorMgr.spawn(NewCronTask("slow", MustNewPattern("* * * * *"), TaskFunc(func() {
		close(start)
		<-finish
	})))
	<-start
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()
	if err := s.Shutdown(ctx); err == nil {
		close(finish)
		t.Fatal("Shutdown should return context timeout")
	}
	close(finish)
	s.Wait()
}
