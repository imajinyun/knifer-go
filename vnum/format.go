package vnum

import numimpl "github.com/imajinyun/knifer-go/internal/num"

func WithFormatFloatFunc(formatter func(float64, byte, int, int) string) FormatOption {
	return numimpl.WithFormatFloatFunc(formatter)
}

func WithFormatIntFunc(formatter func(int64, int) string) FormatOption {
	return numimpl.WithFormatIntFunc(formatter)
}

func DecimalFormat(format string, v float64) string { return numimpl.DecimalFormat(format, v) }

func DecimalFormatWithOptions(format string, v float64, opts ...FormatOption) string {
	return numimpl.DecimalFormatWithOptions(format, v, opts...)
}

func DecimalFormatMoney(v float64) string { return numimpl.DecimalFormatMoney(v) }

func DecimalFormatMoneyWithOptions(v float64, opts ...FormatOption) string {
	return numimpl.DecimalFormatMoneyWithOptions(v, opts...)
}

func FormatPercent(number float64, scale int) string { return numimpl.FormatPercent(number, scale) }

func FormatPercentWithOptions(number float64, scale int, opts ...FormatOption) string {
	return numimpl.FormatPercentWithOptions(number, scale, opts...)
}

func ToStr(number float64) string { return numimpl.ToStr(number) }

func ToStrWithOptions(number float64, opts ...FormatOption) string {
	return numimpl.ToStrWithOptions(number, opts...)
}

func ToStrDefault(number *float64, defaultValue string) string {
	return numimpl.ToStrDefault(number, defaultValue)
}

func ToStrDefaultWithOptions(number *float64, defaultValue string, opts ...FormatOption) string {
	return numimpl.ToStrDefaultWithOptions(number, defaultValue, opts...)
}

func ToStrStrip(number float64, stripTrailingZeros bool) string {
	return numimpl.ToStrStrip(number, stripTrailingZeros)
}

func ToStrStripWithOptions(number float64, stripTrailingZeros bool, opts ...FormatOption) string {
	return numimpl.ToStrStripWithOptions(number, stripTrailingZeros, opts...)
}
