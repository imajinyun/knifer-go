package vnum

import numimpl "github.com/imajinyun/knifer-go/internal/num"

func Round(v float64, scale int) float64 { return numimpl.Round(v, scale) }

func RoundMode(v float64, scale int, mode RoundingMode) float64 {
	return numimpl.RoundMode(v, scale, mode)
}

func RoundStr(v float64, scale int) string { return numimpl.RoundStr(v, scale) }

func RoundStrWithOptions(v float64, scale int, opts ...FormatOption) string {
	return numimpl.RoundStrWithOptions(v, scale, opts...)
}

func RoundHalfEvenFloat(v float64, scale int) float64 { return numimpl.RoundHalfEvenFloat(v, scale) }

func RoundDownFloat(v float64, scale int) float64 { return numimpl.RoundDownFloat(v, scale) }
