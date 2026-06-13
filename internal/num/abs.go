package num

import (
	"fmt"
	"math"

	"github.com/imajinyun/go-knifer/internal/constraint"
)

// AbsInteger returns the absolute value of v, or the zero value when the result
// overflows type T (e.g. abs(math.MinInt8)). Use AbsIntegerE when callers need
// to distinguish overflow from a legitimate zero result.
func AbsInteger[T constraint.Integer](v T) T {
	abs, err := AbsIntegerE(v)
	if err != nil {
		var zero T
		return zero
	}
	return abs
}

// AbsIntegerE returns the absolute value of v.
// It returns an error if the result overflows the type T, which happens only
// for the most negative value of a signed integer type (e.g. math.MinInt8),
// whose absolute value cannot be represented in the same type.
func AbsIntegerE[T constraint.Integer](v T) (T, error) {
	abs := rawAbsInteger(v)
	// Negation overflow: for the minimum signed value, -v wraps back to a
	// negative number, so a non-positive result here signals overflow.
	if v < 0 && abs < 0 {
		return 0, newAbsOverflowError(v)
	}
	return abs, nil
}

// AbsFloat32 returns the absolute value of x.
//
// It clears the sign bit (the highest bit) directly via bit manipulation,
// which avoids a float64 conversion and is faster than math.Abs for float32.
//
// The operator "&^" is bit clear (AND NOT): x &^ y is equivalent to x & (^y).
//
// See:
//   - https://golang.org/ref/spec#Arithmetic_operators
//   - https://yourbasic.org/golang/operators/
func AbsFloat32(x float32) float32 {
	const signBitMask = uint32(1) << 31
	return math.Float32frombits(math.Float32bits(x) &^ signBitMask)
}

// AbsFloat64 returns the absolute value of x.
func AbsFloat64(x float64) float64 {
	return math.Abs(x)
}

// rawAbsInteger returns the absolute value of v without overflow checking.
// For the most negative signed value, the result wraps and stays negative.
func rawAbsInteger[T constraint.Integer](v T) T {
	if v < 0 {
		return -v
	}
	return v
}

// newAbsOverflowError builds the error returned when abs(v) overflows type T.
func newAbsOverflowError[T constraint.Integer](v T) error {
	return fmt.Errorf("%T overflow: abs(%d)", v, v)
}
