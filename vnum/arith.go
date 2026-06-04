package vnum

import (
	"math/big"

	numimpl "github.com/imajinyun/go-knifer/internal/num"
)

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
