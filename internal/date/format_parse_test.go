package date

import (
	"testing"
	"time"
)

func TestFormatAndParse(t *testing.T) {
	tt := time.Date(2024, 7, 15, 10, 20, 30, 0, time.Local)
	if got := FormatDateNorm(tt); got != "2024-07-15 10:20:30" {
		t.Fatalf("FormatDateNorm: %q", got)
	}
	if got := FormatDateOnly(tt); got != "2024-07-15" {
		t.Fatalf("FormatDateOnly: %q", got)
	}
	if got := FormatTimeOnly(tt); got != "10:20:30" {
		t.Fatalf("FormatTimeOnly: %q", got)
	}
	if got := FormatDate(tt, ""); got != "2024-07-15 10:20:30" {
		t.Fatalf("FormatDate empty layout: %q", got)
	}
	if got := FormatDate(tt, "2006/01/02"); got != "2024/07/15" {
		t.Fatalf("FormatDate custom layout: %q", got)
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
