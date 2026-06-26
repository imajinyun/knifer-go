package vnum

import numimpl "github.com/imajinyun/knifer-go/internal/num"

func WithParseIntFunc(parser func(string, int, int) (int64, error)) ParseOption {
	return numimpl.WithParseIntFunc(parser)
}

func WithParseFloatFunc(parser func(string, int) (float64, error)) ParseOption {
	return numimpl.WithParseFloatFunc(parser)
}

func WithDoubleParseFloatFunc(parser func(string, int) (float64, error)) DoubleOption {
	return numimpl.WithDoubleParseFloatFunc(parser)
}

func WithDoubleFormatFloatFunc(formatter func(float64, byte, int, int) string) DoubleOption {
	return numimpl.WithDoubleFormatFloatFunc(formatter)
}

func ParseInt(number string) int { return numimpl.ParseInt(number) }

func ParseIntWithOptions(number string, opts ...ParseOption) int {
	return numimpl.ParseIntWithOptions(number, opts...)
}

func ParseLong(number string) int64 { return numimpl.ParseLong(number) }

func ParseLongWithOptions(number string, opts ...ParseOption) int64 {
	return numimpl.ParseLongWithOptions(number, opts...)
}

func ParseFloat(number string) float32 { return numimpl.ParseFloat(number) }

func ParseFloatWithOptions(number string, opts ...ParseOption) float32 {
	return numimpl.ParseFloatWithOptions(number, opts...)
}

func ParseDouble(number string) float64 { return numimpl.ParseDouble(number) }

func ParseDoubleWithOptions(number string, opts ...ParseOption) float64 {
	return numimpl.ParseDoubleWithOptions(number, opts...)
}

func ParseNumber(numberStr string) (float64, error) { return numimpl.ParseNumber(numberStr) }

func ParseNumberWithOptions(numberStr string, opts ...ParseOption) (float64, error) {
	return numimpl.ParseNumberWithOptions(numberStr, opts...)
}

func ParseIntDefault(numberStr string, defaultValue int) int {
	return numimpl.ParseIntDefault(numberStr, defaultValue)
}

func ParseIntDefaultWithOptions(numberStr string, defaultValue int, opts ...ParseOption) int {
	return numimpl.ParseIntDefaultWithOptions(numberStr, defaultValue, opts...)
}

func ParseLongDefault(numberStr string, defaultValue int64) int64 {
	return numimpl.ParseLongDefault(numberStr, defaultValue)
}

func ParseLongDefaultWithOptions(numberStr string, defaultValue int64, opts ...ParseOption) int64 {
	return numimpl.ParseLongDefaultWithOptions(numberStr, defaultValue, opts...)
}

func ParseFloatDefault(numberStr string, defaultValue float32) float32 {
	return numimpl.ParseFloatDefault(numberStr, defaultValue)
}

func ParseFloatDefaultWithOptions(numberStr string, defaultValue float32, opts ...ParseOption) float32 {
	return numimpl.ParseFloatDefaultWithOptions(numberStr, defaultValue, opts...)
}

func ParseDoubleDefault(numberStr string, defaultValue float64) float64 {
	return numimpl.ParseDoubleDefault(numberStr, defaultValue)
}

func ParseDoubleDefaultWithOptions(numberStr string, defaultValue float64, opts ...ParseOption) float64 {
	return numimpl.ParseDoubleDefaultWithOptions(numberStr, defaultValue, opts...)
}
