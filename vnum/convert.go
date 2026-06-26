package vnum

import (
	"math/big"

	numimpl "github.com/imajinyun/knifer-go/internal/num"
)

func ToBigDecimal(numberStr string) *big.Rat { return numimpl.ToBigDecimal(numberStr) }

func ToBigDecimalWithOptions(numberStr string, opts ...ParseOption) *big.Rat {
	return numimpl.ToBigDecimalWithOptions(numberStr, opts...)
}

func ToBigInteger(number string) *big.Int { return numimpl.ToBigInteger(number) }

func Count(total, part int) int { return numimpl.Count(total, part) }

func Null2Zero(decimal *big.Rat) *big.Rat { return numimpl.Null2Zero(decimal) }

func Zero2One(value int) int { return numimpl.Zero2One(value) }

func NullToZero[T Number](number *T) T { return numimpl.NullToZero(number) }

func NullBigIntToZero(number *big.Int) *big.Int { return numimpl.NullBigIntToZero(number) }

func NullBigDecimalToZero(number *big.Rat) *big.Rat { return numimpl.NullBigDecimalToZero(number) }

func NewBigInteger(str string) (*big.Int, bool) { return numimpl.NewBigInteger(str) }

func PartValue(total, partCount int) int { return numimpl.PartValue(total, partCount) }

func PartValueWithMode(total, partCount int, plusOneWhenHasRem bool) int {
	return numimpl.PartValueWithMode(total, partCount, plusOneWhenHasRem)
}

func ToDouble(value any) float64 { return numimpl.ToDouble(value) }

func ToDoubleWithOptions(value any, opts ...DoubleOption) float64 {
	return numimpl.ToDoubleWithOptions(value, opts...)
}
