// Package date provides date and time helpers.
package date

import (
	"strings"
	"time"
)

// This file provides date and time helpers aligned with the utility toolkit-core DateUtil.

// Common date/time layouts. Go layouts use the reference time 2006-01-02 15:04:05.
const (
	NormPattern         = "2006-01-02 15:04:05"
	NormDatePattern     = "2006-01-02"
	NormTimePattern     = "15:04:05"
	NormDatetimePattern = NormPattern
	PureDatePattern     = "20060102"
	PureDatetimePattern = "20060102150405"
	HTTPPattern         = time.RFC1123
	UTCPattern          = "2006-01-02T15:04:05Z"
)

type parseConfig struct {
	location *time.Location
}

// ParseOption customizes date parsing helpers.
type ParseOption func(*parseConfig)

// WithLocation sets the time zone used when parsing layouts without zone information.
func WithLocation(location *time.Location) ParseOption {
	return func(c *parseConfig) { c.location = location }
}

func applyParseOptions(opts []ParseOption) parseConfig {
	cfg := parseConfig{location: time.Local}
	for _, opt := range opts {
		if opt != nil {
			opt(&cfg)
		}
	}
	if cfg.location == nil {
		cfg.location = time.Local
	}
	return cfg
}

// Now returns the current local time.
func Now() time.Time { return time.Now() }

// Today returns the start of the current day.
func Today() time.Time { return BeginOfDay(time.Now()) }

// FormatDate formats t with layout. An empty layout falls back to NormPattern.
func FormatDate(t time.Time, layout string) string {
	if layout == "" {
		layout = NormPattern
	}
	return t.Format(layout)
}

// FormatDateNorm formats t as yyyy-MM-dd HH:mm:ss.
func FormatDateNorm(t time.Time) string { return t.Format(NormPattern) }

// FormatDateOnly formats t as yyyy-MM-dd.
func FormatDateOnly(t time.Time) string { return t.Format(NormDatePattern) }

// FormatTimeOnly formats t as HH:mm:ss.
func FormatTimeOnly(t time.Time) string { return t.Format(NormTimePattern) }

// ParseDate parses common date/time formats in the local time zone.
func ParseDate(s string) (time.Time, error) {
	return ParseDateWithOptions(s)
}

// ParseDateWithOptions parses common date/time formats with explicit options.
func ParseDateWithOptions(s string, opts ...ParseOption) (time.Time, error) {
	cfg := applyParseOptions(opts)
	s = strings.TrimSpace(s)
	if s == "" {
		return time.Time{}, invalidDateInputf("empty date string")
	}
	patterns := []string{
		NormPattern,
		NormDatePattern,
		NormTimePattern,
		PureDatetimePattern,
		PureDatePattern,
		UTCPattern,
		time.RFC3339,
		time.RFC1123,
		"2006/01/02 15:04:05",
		"2006/01/02",
		"2006-01-02T15:04:05",
	}
	for _, p := range patterns {
		if t, err := time.ParseInLocation(p, s, cfg.location); err == nil {
			return t, nil
		}
	}
	return time.Time{}, invalidDateInputf("unsupported date format: %s", s)
}

// ParseDateLayout parses s with the specified Go layout in the local time zone.
func ParseDateLayout(s, layout string) (time.Time, error) {
	return ParseDateLayoutWithOptions(s, layout)
}

// ParseDateLayoutWithOptions parses s with the specified Go layout and explicit options.
func ParseDateLayoutWithOptions(s, layout string, opts ...ParseOption) (time.Time, error) {
	cfg := applyParseOptions(opts)
	t, err := time.ParseInLocation(layout, s, cfg.location)
	if err != nil {
		return time.Time{}, wrapDateParse("parse date with layout", err)
	}
	return t, nil
}

// BeginOfDay returns midnight at the beginning of t's day.
func BeginOfDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

// EndOfDay returns the last nanosecond of t's day.
func EndOfDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, int(time.Second-time.Nanosecond), t.Location())
}

// BeginOfMonth returns the first instant of t's month.
func BeginOfMonth(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, t.Location())
}

// EndOfMonth returns the last nanosecond of t's month.
func EndOfMonth(t time.Time) time.Time {
	first := BeginOfMonth(t)
	return EndOfDay(first.AddDate(0, 1, -1))
}

// BeginOfYear returns the first instant of t's year.
func BeginOfYear(t time.Time) time.Time {
	return time.Date(t.Year(), 1, 1, 0, 0, 0, 0, t.Location())
}

// EndOfYear returns the last nanosecond of t's year.
func EndOfYear(t time.Time) time.Time {
	return EndOfDay(time.Date(t.Year(), 12, 31, 0, 0, 0, 0, t.Location()))
}

// OffsetDay offsets t by days.
func OffsetDay(t time.Time, days int) time.Time { return t.AddDate(0, 0, days) }

// OffsetMonth offsets t by months.
func OffsetMonth(t time.Time, months int) time.Time { return t.AddDate(0, months, 0) }

// OffsetYear offsets t by years.
func OffsetYear(t time.Time, years int) time.Time { return t.AddDate(years, 0, 0) }

// OffsetHour offsets t by hours.
func OffsetHour(t time.Time, hours int) time.Time { return t.Add(time.Duration(hours) * time.Hour) }

// OffsetMinute offsets t by minutes.
func OffsetMinute(t time.Time, minutes int) time.Time {
	return t.Add(time.Duration(minutes) * time.Minute)
}

// OffsetSecond offsets t by seconds.
func OffsetSecond(t time.Time, seconds int) time.Time {
	return t.Add(time.Duration(seconds) * time.Second)
}

// BetweenDays returns the absolute whole-day distance between two times.
func BetweenDays(a, b time.Time) int {
	d := b.Sub(a) / (24 * time.Hour)
	if d < 0 {
		d = -d
	}
	return int(d)
}

// IsSameDay reports whether two times fall on the same calendar day.
func IsSameDay(a, b time.Time) bool {
	return a.Year() == b.Year() && a.YearDay() == b.YearDay()
}
