package obj

import (
	"fmt"
	"math"
	"math/big"
	"reflect"
	"strings"

	refimpl "github.com/imajinyun/go-knifer/internal/ref"
)

// Ordered is the set of built-in ordered value types.
type Ordered interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr |
		~float32 | ~float64 | ~string
}

// Equal reports whether a and b are equal. Numeric values are compared by value.
func Equal(a, b any) bool {
	if IsNil(a) || IsNil(b) {
		return IsNil(a) && IsNil(b)
	}
	if ar, ok := numberAsRat(a); ok {
		if br, ok := numberAsRat(b); ok {
			return ar.Cmp(br) == 0
		}
	}
	return reflect.DeepEqual(a, b)
}

// Equals is an alias of Equal.
func Equals(a, b any) bool { return Equal(a, b) }

// NotEqual reports whether a and b are not equal.
func NotEqual(a, b any) bool { return !Equal(a, b) }

// Length returns the length of a string, array, slice, map, or channel.
// Nil values return 0 and unsupported values return -1.
func Length(v any) int {
	if IsNil(v) {
		return 0
	}
	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.String, reflect.Array, reflect.Slice, reflect.Map, reflect.Chan:
		return rv.Len()
	default:
		return -1
	}
}

// Contains reports whether obj contains element.
func Contains(obj, element any) bool {
	if IsNil(obj) {
		return false
	}
	rv := reflect.ValueOf(obj)
	switch rv.Kind() {
	case reflect.String:
		if IsNil(element) {
			return false
		}
		return strings.Contains(rv.String(), fmt.Sprint(element))
	case reflect.Map:
		for _, key := range rv.MapKeys() {
			if Equal(rv.MapIndex(key).Interface(), element) {
				return true
			}
		}
		return false
	case reflect.Array, reflect.Slice:
		for i := 0; i < rv.Len(); i++ {
			if Equal(rv.Index(i).Interface(), element) {
				return true
			}
		}
		return false
	default:
		return false
	}
}

// IsNil reports whether v is nil, including typed nil values.
func IsNil(v any) bool { return refimpl.IsNil(v) }

// IsNull is an alias of IsNil.
func IsNull(v any) bool { return IsNil(v) }

// IsNotNil reports whether v is not nil.
func IsNotNil(v any) bool { return !IsNil(v) }

// IsNotNull is an alias of IsNotNil.
func IsNotNull(v any) bool { return IsNotNil(v) }

// IsEmpty reports whether v is nil or an empty string, array, slice, map, or channel.
func IsEmpty(v any) bool {
	if IsNil(v) {
		return true
	}
	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.String, reflect.Array, reflect.Slice, reflect.Map, reflect.Chan:
		return rv.Len() == 0
	default:
		return false
	}
}

// IsNotEmpty reports whether v is not empty.
func IsNotEmpty(v any) bool { return !IsEmpty(v) }

// DefaultIfNil returns defaultValue when object is nil.
func DefaultIfNil[T any](object *T, defaultValue T) T {
	if object == nil {
		return defaultValue
	}
	return *object
}

// DefaultIfNilFunc returns a supplier value when object is nil.
func DefaultIfNilFunc[T any](object *T, supplier func() T) T {
	if object == nil {
		return supplier()
	}
	return *object
}

// DefaultIfNilApply returns defaultValue when source is nil; otherwise it maps source.
func DefaultIfNilApply[T any, R any](source *T, handle func(T) R, defaultValue R) R {
	if source == nil {
		return defaultValue
	}
	return handle(*source)
}

// Apply maps source when it is not nil; otherwise it returns the zero value.
func Apply[T any, R any](source *T, handle func(T) R) R {
	var zero R
	return DefaultIfNilApply(source, handle, zero)
}

// Accept calls consumer when source is not nil.
func Accept[T any](source *T, consumer func(T)) {
	if source != nil {
		consumer(*source)
	}
}

// IsBasicType reports whether object is a built-in scalar type or string.
func IsBasicType(object any) bool {
	if IsNil(object) {
		return false
	}
	switch reflect.TypeOf(object).Kind() {
	case reflect.Bool,
		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr,
		reflect.Float32, reflect.Float64,
		reflect.Complex64, reflect.Complex128,
		reflect.String:
		return true
	default:
		return false
	}
}

// IsValidIfNumber reports false for NaN or infinite float values. Non-number values are valid.
func IsValidIfNumber(object any) bool {
	switch v := object.(type) {
	case float32:
		return !math.IsNaN(float64(v)) && !math.IsInf(float64(v), 0)
	case float64:
		return !math.IsNaN(v) && !math.IsInf(v, 0)
	default:
		return true
	}
}

// Compare compares two ordered values. Nil pointers are ordered after non-nil values by default.
func Compare[T Ordered](a, b *T) int { return CompareNull(a, b, true) }

// CompareNull compares two ordered values and controls nil ordering.
func CompareNull[T Ordered](a, b *T, nilGreater bool) int {
	if a == nil && b == nil {
		return 0
	}
	if a == nil {
		if nilGreater {
			return 1
		}
		return -1
	}
	if b == nil {
		if nilGreater {
			return -1
		}
		return 1
	}
	if *a < *b {
		return -1
	}
	if *a > *b {
		return 1
	}
	return 0
}

// TypeOf returns the reflection type of object, or nil for nil values.
func TypeOf(object any) reflect.Type { return refimpl.TypeOf(object) }

// TypeName returns the full type name of object.
func TypeName(object any) string {
	t := TypeOf(object)
	if t == nil {
		return ""
	}
	return t.String()
}

// ToString converts object to a string. Nil values become "null".
func ToString(object any) string {
	if IsNil(object) {
		return "null"
	}
	return fmt.Sprint(object)
}

// EmptyCount counts nil or empty values.
func EmptyCount(values ...any) int {
	count := 0
	for _, value := range values {
		if IsEmpty(value) {
			count++
		}
	}
	return count
}

// HasNil reports whether any value is nil.
func HasNil(values ...any) bool {
	for _, value := range values {
		if IsNil(value) {
			return true
		}
	}
	return false
}

// HasNull is an alias of HasNil.
func HasNull(values ...any) bool { return HasNil(values...) }

// HasEmpty reports whether any value is nil or empty.
func HasEmpty(values ...any) bool {
	for _, value := range values {
		if IsEmpty(value) {
			return true
		}
	}
	return false
}

// IsAllEmpty reports whether all values are nil or empty.
func IsAllEmpty(values ...any) bool {
	if len(values) == 0 {
		return true
	}
	for _, value := range values {
		if IsNotEmpty(value) {
			return false
		}
	}
	return true
}

// IsAllNotEmpty reports whether all values are not empty.
func IsAllNotEmpty(values ...any) bool {
	if len(values) == 0 {
		return true
	}
	for _, value := range values {
		if IsEmpty(value) {
			return false
		}
	}
	return true
}

func numberAsRat(v any) (*big.Rat, bool) {
	switch n := v.(type) {
	case int:
		return big.NewRat(int64(n), 1), true
	case int8:
		return big.NewRat(int64(n), 1), true
	case int16:
		return big.NewRat(int64(n), 1), true
	case int32:
		return big.NewRat(int64(n), 1), true
	case int64:
		return big.NewRat(n, 1), true
	case uint:
		return uintRat(uint64(n)), true
	case uint8:
		return uintRat(uint64(n)), true
	case uint16:
		return uintRat(uint64(n)), true
	case uint32:
		return uintRat(uint64(n)), true
	case uint64:
		return uintRat(n), true
	case uintptr:
		return uintRat(uint64(n)), true
	case float32:
		if math.IsNaN(float64(n)) || math.IsInf(float64(n), 0) {
			return nil, false
		}
		return floatRat(float64(n))
	case float64:
		if math.IsNaN(n) || math.IsInf(n, 0) {
			return nil, false
		}
		return floatRat(n)
	default:
		return nil, false
	}
}

func uintRat(n uint64) *big.Rat {
	return new(big.Rat).SetInt(new(big.Int).SetUint64(n))
}

func floatRat(n float64) (*big.Rat, bool) {
	r := new(big.Rat).SetFloat64(n)
	if r == nil {
		return nil, false
	}
	return r, true
}
