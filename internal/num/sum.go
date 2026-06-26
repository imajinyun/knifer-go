package num

import "github.com/imajinyun/knifer-go/internal/constraint"

// SumNumber returns the sum of all elements as float64.
func SumNumber[T constraint.Integer | constraint.Float](values ...T) float64 {
	var sum float64
	for _, v := range values {
		sum += float64(v)
	}
	return sum
}
