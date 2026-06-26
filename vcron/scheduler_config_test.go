package vcron_test

import (
	"errors"
	"testing"
	"time"

	"github.com/imajinyun/knifer-go/vcron"
)

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
