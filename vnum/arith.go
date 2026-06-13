package vnum

import (
	"math/big"

	numimpl "github.com/imajinyun/go-knifer/internal/num"
)

type integer interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 | ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr
}

type float interface {
	~float32 | ~float64
}

type number interface {
	integer | float
}

func Add(values ...float64) float64 { return numimpl.Add(values...) }

func AddStr(values ...string) *big.Rat { return numimpl.AddStr(values...) }

func Sub(values ...float64) float64 { return numimpl.Sub(values...) }

func SubStr(values ...string) *big.Rat { return numimpl.SubStr(values...) }

func Mul(values ...float64) float64 { return numimpl.Mul(values...) }

func MulStr(values ...string) *big.Rat { return numimpl.MulStr(values...) }

func Div(a, b float64, scale ...int) float64 {
	if len(scale) == 0 {
		return numimpl.Div(a, b)
	}
	return numimpl.NumberDiv(a, b, scale[0])
}

func DivWithMode(a, b float64, scale int, mode RoundingMode) float64 {
	return numimpl.DivWithMode(a, b, scale, mode)
}

func CeilDiv(v1, v2 int) int { return numimpl.CeilDiv(v1, v2) }

func Pow(number float64, n int) float64 { return numimpl.Pow(number, n) }

func PowWithMode(number float64, n, scale int, mode RoundingMode) float64 {
	return numimpl.PowWithMode(number, n, scale, mode)
}

// AbsInteger returns the absolute value of v, or zero on signed-min overflow.
func AbsInteger[T integer](v T) T { return numimpl.AbsInteger(v) }

// AbsIntegerE returns the absolute value of v, reporting signed-min overflow.
func AbsIntegerE[T integer](v T) (T, error) { return numimpl.AbsIntegerE(v) }

// AbsFloat32 returns the absolute value of x without widening to float64.
func AbsFloat32(x float32) float32 { return numimpl.AbsFloat32(x) }

// AbsFloat64 returns the absolute value of x.
func AbsFloat64(x float64) float64 { return numimpl.AbsFloat64(x) }

// SumNumber returns the sum of all integer or floating-point values as float64.
func SumNumber[T number](values ...T) float64 { return numimpl.SumNumber(values...) }

// AvgNumber returns the arithmetic mean of all values, or 0 for empty input.
func AvgNumber[T number](values ...T) float64 { return numimpl.AvgNumber(values...) }

// MinInteger returns the smaller of a or b.
func MinInteger[T integer](a, b T) T { return numimpl.MinInteger(a, b) }

// MinIntegers returns the smallest integer value, or zero for empty input.
func MinIntegers[T integer](values ...T) T { return numimpl.MinIntegers(values...) }

// MinFloat64 returns the smaller of a or b.
func MinFloat64(a, b float64) float64 { return numimpl.MinFloat64(a, b) }

// MinFloat64s returns the smallest float64 value, or 0 for empty input.
func MinFloat64s(values ...float64) float64 { return numimpl.MinFloat64s(values...) }

// MaxInteger returns the larger of a or b.
func MaxInteger[T integer](a, b T) T { return numimpl.MaxInteger(a, b) }

// MaxIntegers returns the largest integer value, or zero for empty input.
func MaxIntegers[T integer](values ...T) T { return numimpl.MaxIntegers(values...) }

// MaxFloat64 returns the larger of a or b.
func MaxFloat64(a, b float64) float64 { return numimpl.MaxFloat64(a, b) }

// MaxFloat64s returns the largest float64 value, or 0 for empty input.
func MaxFloat64s(values ...float64) float64 { return numimpl.MaxFloat64s(values...) }
