package num

import "github.com/imajinyun/go-knifer/internal/constraint"

// AvgNumber returns the arithmetic mean of all elements, or 0 for empty input.
func AvgNumber[T constraint.Integer | constraint.Float](values ...T) float64 {
	if len(values) == 0 {
		return 0
	}
	return SumNumber(values...) / float64(len(values))
}
