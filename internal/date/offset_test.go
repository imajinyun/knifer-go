package date

import (
	"testing"
	"time"
)

func TestOffsets(t *testing.T) {
	tt := time.Date(2024, 7, 15, 10, 30, 45, 0, time.Local)
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
	if OffsetMinute(tt, 15).Minute() != 45 {
		t.Fatalf("OffsetMinute failed")
	}
	if OffsetSecond(tt, 30).Second() != 15 {
		t.Fatalf("OffsetSecond failed")
	}
}
