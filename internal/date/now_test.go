package date

import (
	"testing"
	"time"
)

func TestNowAndTodayDefaults(t *testing.T) {
	// Now() and Today() should return non-zero time (within reasonable window)
	now := Now()
	if now.IsZero() {
		t.Fatal("Now() returned zero time")
	}
	today := Today()
	if today.IsZero() {
		t.Fatal("Today() returned zero time")
	}
	// Today should be at the start of the day
	if today.Hour() != 0 || today.Minute() != 0 || today.Second() != 0 {
		t.Fatalf("Today() = %v, should be start of day", today)
	}
}

func TestNowWithOptionsClock(t *testing.T) {
	fixed := time.Date(2026, 6, 6, 12, 34, 56, 0, time.FixedZone("fixed", 8*60*60))
	if got := NowWithOptions(WithClock(func() time.Time { return fixed })); !got.Equal(fixed) {
		t.Fatalf("NowWithOptions = %v, want %v", got, fixed)
	}
	today := TodayWithOptions(WithClock(func() time.Time { return fixed }))
	if !today.Equal(time.Date(2026, 6, 6, 0, 0, 0, 0, fixed.Location())) {
		t.Fatalf("TodayWithOptions = %v", today)
	}
	if got := NowWithOptions(WithClock(nil)); got.IsZero() {
		t.Fatal("NowWithOptions nil clock should fall back to time.Now")
	}
	if got := NowWithOptions(WithClock(func() time.Time { return fixed }), WithClock(nil)); !got.Equal(fixed) {
		t.Fatalf("NowWithOptions nil overwrite clock = %v, want %v", got, fixed)
	}
}
