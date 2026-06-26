package vcron_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/imajinyun/knifer-go/vcron"
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

func TestFacadePatternParserAndErrors(t *testing.T) {
	parseErr := errors.New("bad integer")
	s := vcron.NewSchedulerWithOptions(vcron.WithSchedulerPatternOptions(
		vcron.WithPatternIntParser(func(s string) (int, error) {
			if s == "nope" {
				return 0, parseErr
			}
			return 1, nil
		}),
	))
	err := s.ScheduleWithID("bad", "nope * * * * *", vcron.TaskFunc(func() {}))
	if err == nil || !strings.Contains(err.Error(), "invalid number") || !strings.Contains(err.Error(), "nope") {
		t.Fatalf("ScheduleWithID parser error = %v, want invalid number for nope", err)
	}

	cause := errors.New("cause")
	wrapped := vcron.WrapCronError(cause, "schedule %s", "failed")
	if !errors.Is(wrapped, cause) || !strings.Contains(wrapped.Error(), "schedule failed") {
		t.Fatalf("WrapCronError = %v", wrapped)
	}
	if got := vcron.NewCronError("plain %d", 1); !strings.Contains(got.Error(), "plain 1") {
		t.Fatalf("NewCronError = %v", got)
	}
}
