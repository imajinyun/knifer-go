package cron

import (
	"strconv"
	"testing"
	"time"
)

func TestPatternInvalid(t *testing.T) {
	cases := []string{
		"",
		"* *",
		"60 * * * *",
		"* * * 13 *",
		"* * * * 8",
	}
	for _, c := range cases {
		if _, err := NewPattern(c); err == nil {
			t.Fatalf("expected error for %q", c)
		}
	}
}

func TestNewPatternWithOptionsUsesParser(t *testing.T) {
	parseCalls := 0
	p, err := NewPatternWithOptions("custom * * * *", WithPatternIntParser(func(text string) (int, error) {
		parseCalls++
		if text == "custom" {
			return 30, nil
		}
		return strconv.Atoi(text)
	}))
	if err != nil {
		t.Fatalf("NewPatternWithOptions() error = %v", err)
	}
	if parseCalls == 0 {
		t.Fatal("custom pattern int parser was not called")
	}
	yes := time.Date(2024, 1, 1, 12, 30, 0, 0, time.UTC)
	no := time.Date(2024, 1, 1, 12, 31, 0, 0, time.UTC)
	if !p.Match(yes, false) || p.Match(no, false) {
		t.Fatalf("custom parsed pattern mismatch")
	}
}

func TestSchedulerPatternOptions(t *testing.T) {
	parseCalls := 0
	s := NewSchedulerWithOptions(
		WithIDGenerator(func() string { return "custom-pattern-id" }),
		WithSchedulerPatternOptions(WithPatternIntParser(func(text string) (int, error) {
			parseCalls++
			if text == "custom" {
				return 30, nil
			}
			return strconv.Atoi(text)
		})),
	)
	id, err := s.ScheduleFunc("custom * * * *", func() {})
	if err != nil {
		t.Fatalf("ScheduleFunc with pattern options: %v", err)
	}
	if id != "custom-pattern-id" || parseCalls == 0 {
		t.Fatalf("pattern options not used: id=%q parseCalls=%d", id, parseCalls)
	}
	if got := s.GetPattern(id); got == nil || !got.Match(time.Date(2024, 1, 1, 12, 30, 0, 0, time.UTC), false) {
		t.Fatalf("stored custom pattern = %#v", got)
	}
}
