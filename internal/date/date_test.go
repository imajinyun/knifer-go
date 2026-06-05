package date

import (
	"errors"
	"testing"
	"time"

	knifer "github.com/imajinyun/go-knifer"
)

// Tests cover the utility toolkit-core DateUtilTest.

func TestFormatAndParse(t *testing.T) {
	tt := time.Date(2024, 7, 15, 10, 20, 30, 0, time.Local)
	if got := FormatDateNorm(tt); got != "2024-07-15 10:20:30" {
		t.Fatalf("FormatDateNorm: %q", got)
	}
	if got := FormatDateOnly(tt); got != "2024-07-15" {
		t.Fatalf("FormatDateOnly: %q", got)
	}
	parsed, err := ParseDate("2024-07-15 10:20:30")
	if err != nil {
		t.Fatalf("ParseDate err: %v", err)
	}
	if !parsed.Equal(tt) {
		t.Fatalf("Parsed mismatch: %v", parsed)
	}
	if _, err := ParseDate("2024/07/15"); err != nil {
		t.Fatalf("ParseDate slash: %v", err)
	}
	if _, err := ParseDate("20240715"); err != nil {
		t.Fatalf("ParseDate pure: %v", err)
	}
}

func TestParseDateWithOptionsLocation(t *testing.T) {
	loc := time.FixedZone("biz", 8*60*60)
	parsed, err := ParseDateWithOptions("2024-07-15 10:20:30", WithLocation(loc))
	if err != nil {
		t.Fatalf("ParseDateWithOptions err: %v", err)
	}
	if parsed.Location() != loc || parsed.Format(NormPattern) != "2024-07-15 10:20:30" {
		t.Fatalf("ParseDateWithOptions location = %v, %s", parsed.Location(), parsed.Format(NormPattern))
	}

	parsed, err = ParseDateLayoutWithOptions("2024/07/15 10:20:30", "2006/01/02 15:04:05", WithLocation(loc))
	if err != nil {
		t.Fatalf("ParseDateLayoutWithOptions err: %v", err)
	}
	if parsed.Location() != loc || parsed.Format(NormPattern) != "2024-07-15 10:20:30" {
		t.Fatalf("ParseDateLayoutWithOptions location = %v, %s", parsed.Location(), parsed.Format(NormPattern))
	}
}

func TestBeginEndOf(t *testing.T) {
	tt := time.Date(2024, 7, 15, 10, 20, 30, 123, time.Local)
	if FormatDateNorm(BeginOfDay(tt)) != "2024-07-15 00:00:00" {
		t.Fatalf("BeginOfDay failed")
	}
	if FormatDateOnly(EndOfDay(tt)) != "2024-07-15" || EndOfDay(tt).Hour() != 23 {
		t.Fatalf("EndOfDay failed")
	}
	if FormatDateNorm(BeginOfMonth(tt)) != "2024-07-01 00:00:00" {
		t.Fatalf("BeginOfMonth failed")
	}
	if FormatDateOnly(EndOfMonth(tt)) != "2024-07-31" {
		t.Fatalf("EndOfMonth failed")
	}
	if FormatDateNorm(BeginOfYear(tt)) != "2024-01-01 00:00:00" {
		t.Fatalf("BeginOfYear failed")
	}
	if FormatDateOnly(EndOfYear(tt)) != "2024-12-31" {
		t.Fatalf("EndOfYear failed")
	}
}

func TestOffsets(t *testing.T) {
	tt := time.Date(2024, 7, 15, 10, 0, 0, 0, time.Local)
	if FormatDateOnly(OffsetDay(tt, 1)) != "2024-07-16" {
		t.Fatalf("OffsetDay failed")
	}
	if FormatDateOnly(OffsetMonth(tt, 1)) != "2024-08-15" {
		t.Fatalf("OffsetMonth failed")
	}
	if FormatDateOnly(OffsetYear(tt, -1)) != "2023-07-15" {
		t.Fatalf("OffsetYear failed")
	}
	if OffsetHour(tt, 2).Hour() != 12 {
		t.Fatalf("OffsetHour failed")
	}
}

func TestBetweenAndSameDay(t *testing.T) {
	a := time.Date(2024, 7, 15, 0, 0, 0, 0, time.Local)
	b := time.Date(2024, 7, 20, 0, 0, 0, 0, time.Local)
	if BetweenDays(a, b) != 5 {
		t.Fatalf("BetweenDays failed")
	}
	if !IsSameDay(a, time.Date(2024, 7, 15, 23, 0, 0, 0, time.Local)) {
		t.Fatalf("IsSameDay failed")
	}
}

func TestDateErrorContract(t *testing.T) {
	_, err := ParseDate("")
	assertDateCode(t, err, knifer.ErrCodeInvalidInput)

	_, err = ParseDate("not-a-date")
	assertDateCode(t, err, knifer.ErrCodeInvalidInput)

	_, err = ParseDateLayout("2026-06-05", "bad-layout")
	assertDateCode(t, err, knifer.ErrCodeInvalidInput)
}

func assertDateCode(t *testing.T, err error, code knifer.ErrCode) {
	t.Helper()
	if err == nil {
		t.Fatalf("err = nil, want %s", code)
	}
	if !errors.Is(err, code) {
		t.Fatalf("errors.Is(%v, %s) = false", err, code)
	}
	got, ok := knifer.CodeOf(err)
	if !ok || got != code {
		t.Fatalf("CodeOf(%v) = %q, %v; want %q, true", err, got, ok, code)
	}
	var dateErr *DateError
	if !errors.As(err, &dateErr) {
		t.Fatalf("errors.As(err, *DateError) = false: %v", err)
	}
}
