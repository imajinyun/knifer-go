package vnum

import numimpl "github.com/imajinyun/go-knifer/internal/num"

func ParseInt(number string) int { return numimpl.ParseInt(number) }

func ParseLong(number string) int64 { return numimpl.ParseLong(number) }

func ParseFloat(number string) float32 { return numimpl.ParseFloat(number) }

func ParseDouble(number string) float64 { return numimpl.ParseDouble(number) }

func ParseNumber(numberStr string) (float64, error) { return numimpl.ParseNumber(numberStr) }

func ParseIntDefault(numberStr string, defaultValue int) int {
	return numimpl.ParseIntDefault(numberStr, defaultValue)
}

func ParseLongDefault(numberStr string, defaultValue int64) int64 {
	return numimpl.ParseLongDefault(numberStr, defaultValue)
}

func ParseFloatDefault(numberStr string, defaultValue float32) float32 {
	return numimpl.ParseFloatDefault(numberStr, defaultValue)
}

func ParseDoubleDefault(numberStr string, defaultValue float64) float64 {
	return numimpl.ParseDoubleDefault(numberStr, defaultValue)
}
