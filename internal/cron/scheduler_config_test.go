package cron

import (
	"errors"
	"testing"
	"time"

	knifer "github.com/imajinyun/go-knifer"
)

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
