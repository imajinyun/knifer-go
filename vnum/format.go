package vnum

import numimpl "github.com/imajinyun/go-knifer/internal/num"

func DecimalFormat(format string, v float64) string { return numimpl.DecimalFormat(format, v) }

func DecimalFormatMoney(v float64) string { return numimpl.DecimalFormatMoney(v) }

func FormatPercent(number float64, scale int) string { return numimpl.FormatPercent(number, scale) }

func ToStr(number float64) string { return numimpl.ToStr(number) }

func ToStrDefault(number *float64, defaultValue string) string {
	return numimpl.ToStrDefault(number, defaultValue)
}

func ToStrStrip(number float64, stripTrailingZeros bool) string {
	return numimpl.ToStrStrip(number, stripTrailingZeros)
}
