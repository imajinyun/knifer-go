package errx

import (
	"errors"
	"testing"
)

func TestCollectorGoRunAggregatesConcurrentErrors(t *testing.T) {
	silenceLogrus(t)

	errA := errors.New("a")
	errB := errors.New("b")
	c := NewCollector()
	c.GoRun(func() error { return errA }, "job a")
	c.GoRun(func() error { return errB }, "job b")

	got := c.Error()
	if !ErrorIs(got, errA) || !ErrorIs(got, errB) {
		t.Fatalf("Collector.Error() = %v, want both errors", got)
	}
}

func TestCollectorGoRunUsesRunner(t *testing.T) {
	silenceLogrus(t)

	runnerCalls := 0
	c := NewCollectorWithOptions(WithCollectorRunner(func(fn func()) {
		runnerCalls++
		fn()
	}))
	c.GoRun(func() error { return nil }, "sync")
	if runnerCalls != 1 {
		t.Fatalf("runner calls = %d, want 1", runnerCalls)
	}
	if err := c.Error(); err != nil {
		t.Fatalf("Collector.Error() = %v, want nil", err)
	}
}
