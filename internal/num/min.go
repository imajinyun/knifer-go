package num

import (
	"math"

	"github.com/imajinyun/go-knifer/internal/constraint"
)

// MinInteger returns the smaller of a or b.
func MinInteger[T constraint.Integer](a, b T) T {
	if a <= b {
		return a
	}
	return b
}

// MinIntegers returns the smallest of the given elements, or the zero value when
// no values are provided.
func MinIntegers[T constraint.Integer](values ...T) T {
	if len(values) == 0 {
		var zero T
		return zero
	}
	min := values[0]
	for _, v := range values[1:] {
		if v < min {
			min = v
		}
	}
	return min
}

// MinFloat64 returns the smaller of a or b.
func MinFloat64(a, b float64) float64 {
	return math.Min(a, b)
}

// MinFloat64s returns the smallest of the given elements, or 0 for empty input.
//
// Note: math.Min propagates NaN, so if any element is NaN the result is NaN.
func MinFloat64s(values ...float64) float64 {
	if len(values) == 0 {
		return 0
	}
	min := values[0]
	for _, v := range values[1:] {
		min = math.Min(min, v)
	}
	return min
}
