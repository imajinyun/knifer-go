package vnum

import (
	"math/big"

	numimpl "github.com/imajinyun/go-knifer/internal/num"
)

func Compare[T Ordered](x, y T) int { return numimpl.Compare(x, y) }

func IsGreater[T Ordered](a, b T) bool { return numimpl.IsGreater(a, b) }

func IsGreaterOrEqual[T Ordered](a, b T) bool { return numimpl.IsGreaterOrEqual(a, b) }

func IsLess[T Ordered](a, b T) bool { return numimpl.IsLess(a, b) }

func IsLessOrEqual[T Ordered](a, b T) bool { return numimpl.IsLessOrEqual(a, b) }

func IsIn[T Ordered](value, minInclude, maxInclude T) bool {
	return numimpl.IsIn(value, minInclude, maxInclude)
}

func Equals(a, b float64) bool { return numimpl.Equals(a, b) }

func EqualsExact(a, b float64) bool { return numimpl.EqualsExact(a, b) }

func EqualsFloat32Exact(a, b float32) bool { return numimpl.EqualsFloat32Exact(a, b) }

func EqualsInt64(a, b int64) bool { return numimpl.EqualsInt64(a, b) }

func EqualsBigDecimal(a, b *big.Rat) bool { return numimpl.EqualsBigDecimal(a, b) }

func EqualsChar(c1, c2 rune, ignoreCase bool) bool { return numimpl.EqualsChar(c1, c2, ignoreCase) }

func Min[T Ordered](nums ...T) T { return numimpl.Min(nums...) }

func Max[T Ordered](nums ...T) T { return numimpl.Max(nums...) }

func Sum[T Number](nums ...T) T { return numimpl.Sum(nums...) }

func Avg[T Number](nums ...T) float64 { return numimpl.Avg(nums...) }

func IsBeside[T ~int | ~int64](number1, number2 T) bool { return numimpl.IsBeside(number1, number2) }
