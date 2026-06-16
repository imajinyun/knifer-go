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

func TestDateFacadeClockOptions(t *testing.T) {
	fixed := time.Date(2026, 6, 6, 12, 34, 56, 0, time.FixedZone("facade-clock", 8*60*60))
	if got := Now(); got.IsZero() {
		t.Fatal("Now() returned zero time")
	}
	if got := Today(); got.Hour() != 0 || got.Minute() != 0 || got.Second() != 0 || got.Nanosecond() != 0 {
		t.Fatalf("Today() = %v, want beginning of local day", got)
	}
	if got := NowWithOptions(WithClock(func() time.Time { return fixed })); !got.Equal(fixed) {
		t.Fatalf("NowWithOptions = %v, want %v", got, fixed)
	}
	if got := TodayWithOptions(WithClock(func() time.Time { return fixed })); !got.Equal(time.Date(2026, 6, 6, 0, 0, 0, 0, fixed.Location())) {
		t.Fatalf("TodayWithOptions = %v", got)
	}
}

func TestDateFacadeParseProviderOptions(t *testing.T) {
	loc := time.FixedZone("parse-provider", 9*60*60)
	want := time.Date(2026, 6, 16, 10, 11, 12, 0, loc)
	called := false
	got, err := ParseLayoutWithOptions("custom", "layout",
		WithLocation(loc),
		WithParseInLocationFunc(func(layout, value string, location *time.Location) (time.Time, error) {
			called = true
			if layout != "layout" || value != "custom" || location != loc {
				t.Fatalf("parser args layout=%q value=%q location=%v", layout, value, location)
			}
			return want, nil
		}),
	)
	if err != nil || !got.Equal(want) || !called {
		t.Fatalf("ParseLayoutWithOptions custom parser = %v, %v called=%v", got, err, called)
	}

	got, err = ParseWithOptions("custom",
		WithLocation(loc),
		WithParseInLocationFunc(func(layout, value string, location *time.Location) (time.Time, error) {
			if layout != NormDatetimePattern || value != "custom" || location != loc {
				t.Fatalf("ParseWithOptions parser args layout=%q value=%q location=%v", layout, value, location)
			}
			return want, nil
		}),
	)
	if err != nil || !got.Equal(want) {
		t.Fatalf("ParseWithOptions custom parser = %v, %v", got, err)
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
