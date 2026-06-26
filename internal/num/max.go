package num

import (
	"math"

	"github.com/imajinyun/knifer-go/internal/constraint"
)

// MaxInteger returns the larger of a or b.
func MaxInteger[T constraint.Integer](a, b T) T {
	if a >= b {
		return a
	}
	return b
}

// MaxIntegers returns the largest of the given elements, or the zero value when
// no values are provided.
func MaxIntegers[T constraint.Integer](values ...T) T {
	if len(values) == 0 {
		var zero T
		return zero
	}
	max := values[0]
	for _, v := range values[1:] {
		if v > max {
			max = v
		}
	}
	return max
}

// MaxFloat64 returns the larger of a or b.
func MaxFloat64(a, b float64) float64 {
	return math.Max(a, b)
}

// MaxFloat64s returns the largest of the given elements, or 0 for empty input.
//
// Note: math.Max propagates NaN, so if any element is NaN the result is NaN.
func MaxFloat64s(values ...float64) float64 {
	if len(values) == 0 {
		return 0
	}
	max := values[0]
	for _, v := range values[1:] {
		max = math.Max(max, v)
	}
	return max
}
