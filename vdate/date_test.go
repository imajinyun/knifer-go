package vdate

import (
	"errors"
	"testing"
	"time"

	knifer "github.com/imajinyun/go-knifer"
)

func TestDateFacade(t *testing.T) {
	base := time.Date(2026, 5, 30, 1, 2, 3, 0, time.Local)
	if Format(base, "") != "2026-05-30 01:02:03" || FormatNorm(base) != "2026-05-30 01:02:03" {
		t.Fatal("format failed")
	}
	if FormatDateOnly(base) != "2026-05-30" || FormatTimeOnly(base) != "01:02:03" {
		t.Fatal("date/time format failed")
	}
	if got, err := Parse("2026-05-30"); err != nil || got.Year() != 2026 {
		t.Fatalf("Parse = %v, %v", got, err)
	}
	if got, err := ParseLayout("2026/05/30", "2006/01/02"); err != nil || got.Day() != 30 {
		t.Fatalf("ParseLayout = %v, %v", got, err)
	}
	loc := time.FixedZone("facade", 8*60*60)
	if got, err := ParseWithOptions("2026-05-30", WithLocation(loc)); err != nil || got.Location() != loc {
		t.Fatalf("ParseWithOptions = %v, %v", got, err)
	}
	if got, err := ParseLayoutWithOptions("2026/05/30", "2006/01/02", WithLocation(loc)); err != nil || got.Location() != loc {
		t.Fatalf("ParseLayoutWithOptions = %v, %v", got, err)
	}
	if BeginOfDay(base).Hour() != 0 || EndOfDay(base).Hour() != 23 {
		t.Fatal("begin/end day failed")
	}
	if BeginOfMonth(base).Day() != 1 || EndOfMonth(base).Day() != 31 {
		t.Fatal("begin/end month failed")
	}
	if BeginOfYear(base).Month() != time.January || EndOfYear(base).Month() != time.December {
		t.Fatal("begin/end year failed")
	}
	if OffsetDay(base, 1).Day() != 31 || OffsetMonth(base, 1).Month() != time.June || OffsetYear(base, 1).Year() != 2027 {
		t.Fatal("date offset failed")
	}
	if OffsetHour(base, 1).Hour() != 2 || OffsetMinute(base, 1).Minute() != 3 || OffsetSecond(base, 1).Second() != 4 {
		t.Fatal("time offset failed")
	}
	if BetweenDays(base, base.Add(48*time.Hour)) != 2 || !IsSameDay(base, base.Add(time.Hour)) {
		t.Fatal("comparison failed")
	}
}

func TestDateFacadeErrorContract(t *testing.T) {
	_, err := Parse("")
	assertFacadeDateCode(t, err, knifer.ErrCodeInvalidInput)

	_, err = Parse("not-a-date")
	assertFacadeDateCode(t, err, knifer.ErrCodeInvalidInput)

	_, err = ParseLayout("2026-06-05", "bad-layout")
	assertFacadeDateCode(t, err, knifer.ErrCodeInvalidInput)
}

func assertFacadeDateCode(t *testing.T, err error, code knifer.ErrCode) {
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
	var dateErr *Error
	if !errors.As(err, &dateErr) {
		t.Fatalf("errors.As(err, *vdate.Error) = false: %v", err)
	}
}
