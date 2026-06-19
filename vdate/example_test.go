package vdate_test

import (
	"fmt"
	"time"

	"github.com/imajinyun/go-knifer/vdate"
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
