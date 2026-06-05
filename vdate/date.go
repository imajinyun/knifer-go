package vdate

import (
	"time"

	dateimpl "github.com/imajinyun/go-knifer/internal/date"
)

// Error is the date module error type.
type Error = dateimpl.DateError

// ParseOption customizes date parsing helpers.
type ParseOption = dateimpl.ParseOption

const (
	NormPattern         = dateimpl.NormPattern
	NormDatePattern     = dateimpl.NormDatePattern
	NormTimePattern     = dateimpl.NormTimePattern
	NormDatetimePattern = dateimpl.NormDatetimePattern
	PureDatePattern     = dateimpl.PureDatePattern
	PureDatetimePattern = dateimpl.PureDatetimePattern
	HTTPPattern         = dateimpl.HTTPPattern
	UTCPattern          = dateimpl.UTCPattern
)

func Now() time.Time                                   { return dateimpl.Now() }
func Today() time.Time                                 { return dateimpl.Today() }
func Format(t time.Time, layout string) string         { return dateimpl.FormatDate(t, layout) }
func FormatNorm(t time.Time) string                    { return dateimpl.FormatDateNorm(t) }
func FormatDateOnly(t time.Time) string                { return dateimpl.FormatDateOnly(t) }
func FormatTimeOnly(t time.Time) string                { return dateimpl.FormatTimeOnly(t) }
func Parse(s string) (time.Time, error)                { return dateimpl.ParseDate(s) }
func ParseLayout(s, layout string) (time.Time, error)  { return dateimpl.ParseDateLayout(s, layout) }
func WithLocation(location *time.Location) ParseOption { return dateimpl.WithLocation(location) }
func ParseWithOptions(s string, opts ...ParseOption) (time.Time, error) {
	return dateimpl.ParseDateWithOptions(s, opts...)
}

func ParseLayoutWithOptions(s, layout string, opts ...ParseOption) (time.Time, error) {
	return dateimpl.ParseDateLayoutWithOptions(s, layout, opts...)
}

func BeginOfDay(t time.Time) time.Time                { return dateimpl.BeginOfDay(t) }
func EndOfDay(t time.Time) time.Time                  { return dateimpl.EndOfDay(t) }
func BeginOfMonth(t time.Time) time.Time              { return dateimpl.BeginOfMonth(t) }
func EndOfMonth(t time.Time) time.Time                { return dateimpl.EndOfMonth(t) }
func BeginOfYear(t time.Time) time.Time               { return dateimpl.BeginOfYear(t) }
func EndOfYear(t time.Time) time.Time                 { return dateimpl.EndOfYear(t) }
func OffsetDay(t time.Time, days int) time.Time       { return dateimpl.OffsetDay(t, days) }
func OffsetMonth(t time.Time, months int) time.Time   { return dateimpl.OffsetMonth(t, months) }
func OffsetYear(t time.Time, years int) time.Time     { return dateimpl.OffsetYear(t, years) }
func OffsetHour(t time.Time, hours int) time.Time     { return dateimpl.OffsetHour(t, hours) }
func OffsetMinute(t time.Time, minutes int) time.Time { return dateimpl.OffsetMinute(t, minutes) }
func OffsetSecond(t time.Time, seconds int) time.Time { return dateimpl.OffsetSecond(t, seconds) }
func BetweenDays(a, b time.Time) int                  { return dateimpl.BetweenDays(a, b) }
func IsSameDay(a, b time.Time) bool                   { return dateimpl.IsSameDay(a, b) }
