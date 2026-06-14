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
