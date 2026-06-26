package vdate_test

import (
	"fmt"
	"time"

	"github.com/imajinyun/knifer-go/vdate"
)

func ExampleFormatNorm() {
	t := time.Date(2026, 6, 2, 15, 4, 5, 0, time.UTC)
	fmt.Println(vdate.FormatNorm(t))
	// Output: 2026-06-02 15:04:05
}

func ExampleBeginOfDay() {
	t := time.Date(2026, 6, 2, 15, 4, 5, 0, time.UTC)
	fmt.Println(vdate.FormatNorm(vdate.BeginOfDay(t)))
	// Output: 2026-06-02 00:00:00
}

func ExampleOffsetDay() {
	t := time.Date(2026, 6, 2, 0, 0, 0, 0, time.UTC)
	fmt.Println(vdate.FormatDateOnly(vdate.OffsetDay(t, 3)))
	// Output: 2026-06-05
}

func ExampleBetweenDays() {
	a := time.Date(2026, 6, 2, 0, 0, 0, 0, time.UTC)
	b := time.Date(2026, 6, 5, 0, 0, 0, 0, time.UTC)
	fmt.Println(vdate.BetweenDays(a, b))
	// Output: 3
}

func ExampleEndOfMonth() {
	t := time.Date(2026, 6, 2, 15, 4, 5, 0, time.UTC)

	fmt.Println(vdate.FormatNorm(vdate.EndOfMonth(t)))
	// Output: 2026-06-30 23:59:59
}

func ExampleParse() {
	t, err := vdate.Parse("2026-06-02 15:04:05")
	fmt.Println(vdate.FormatNorm(t))
	fmt.Println(err)
	// Output:
	// 2026-06-02 15:04:05
	// <nil>
}

func ExampleParseLayout() {
	t, err := vdate.ParseLayout("02/06/2026", "02/01/2006")
	fmt.Println(vdate.FormatDateOnly(t))
	fmt.Println(err)
	// Output:
	// 2026-06-02
	// <nil>
}

func ExampleFormat() {
	t := time.Date(2026, 6, 2, 15, 4, 5, 0, time.UTC)

	fmt.Println(vdate.Format(t, "2006/01/02"))
	fmt.Println(vdate.FormatTimeOnly(t))
	// Output:
	// 2026/06/02
	// 15:04:05
}

func ExampleTodayWithOptions() {
	clock := func() time.Time {
		return time.Date(2026, 6, 2, 15, 4, 5, 0, time.UTC)
	}

	fmt.Println(vdate.FormatNorm(vdate.TodayWithOptions(vdate.WithClock(clock))))
	// Output: 2026-06-02 00:00:00
}

func ExampleEndOfDay() {
	t := time.Date(2026, 6, 2, 15, 4, 5, 0, time.UTC)

	fmt.Println(vdate.FormatNorm(vdate.EndOfDay(t)))
	// Output: 2026-06-02 23:59:59
}

func ExampleBeginOfMonth() {
	t := time.Date(2026, 6, 2, 15, 4, 5, 0, time.UTC)

	fmt.Println(vdate.FormatNorm(vdate.BeginOfMonth(t)))
	// Output: 2026-06-01 00:00:00
}

func ExampleBeginOfYear() {
	t := time.Date(2026, 6, 2, 15, 4, 5, 0, time.UTC)

	fmt.Println(vdate.FormatNorm(vdate.BeginOfYear(t)))
	// Output: 2026-01-01 00:00:00
}

func ExampleEndOfYear() {
	t := time.Date(2026, 6, 2, 15, 4, 5, 0, time.UTC)

	fmt.Println(vdate.FormatNorm(vdate.EndOfYear(t)))
	// Output: 2026-12-31 23:59:59
}

func ExampleOffsetMonth() {
	t := time.Date(2026, 6, 2, 0, 0, 0, 0, time.UTC)

	fmt.Println(vdate.FormatDateOnly(vdate.OffsetMonth(t, -1)))
	// Output: 2026-05-02
}

func ExampleOffsetHour() {
	t := time.Date(2026, 6, 2, 15, 4, 5, 0, time.UTC)

	fmt.Println(vdate.FormatNorm(vdate.OffsetHour(t, 2)))
	// Output: 2026-06-02 17:04:05
}

func ExampleIsSameDay() {
	a := time.Date(2026, 6, 2, 0, 0, 0, 0, time.UTC)
	b := time.Date(2026, 6, 2, 23, 59, 59, 0, time.UTC)
	c := time.Date(2026, 6, 3, 0, 0, 0, 0, time.UTC)

	fmt.Println(vdate.IsSameDay(a, b))
	fmt.Println(vdate.IsSameDay(a, c))
	// Output:
	// true
	// false
}
