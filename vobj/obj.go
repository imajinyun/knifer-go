package vobj

import (
	"reflect"

	objimpl "github.com/imajinyun/go-knifer/internal/obj"
)

// Ordered is the set of built-in ordered value types.
type Ordered interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr |
		~float32 | ~float64 | ~string
}

// Equal reports whether a and b are equal. Numeric values are compared by value.
func Equal(a, b any) bool { return objimpl.Equal(a, b) }

// Equals is an alias of Equal.
func Equals(a, b any) bool { return objimpl.Equals(a, b) }

// NotEqual reports whether a and b are not equal.
func NotEqual(a, b any) bool { return objimpl.NotEqual(a, b) }

// Length returns the length of a string, array, slice, map, or channel.
func Length(v any) int { return objimpl.Length(v) }

// Contains reports whether obj contains element.
func Contains(obj, element any) bool { return objimpl.Contains(obj, element) }

// IsNil reports whether v is nil, including typed nil values.
func IsNil(v any) bool { return objimpl.IsNil(v) }

// IsNull is an alias of IsNil.
func IsNull(v any) bool { return objimpl.IsNull(v) }

// IsNotNil reports whether v is not nil.
func IsNotNil(v any) bool { return objimpl.IsNotNil(v) }

// IsNotNull is an alias of IsNotNil.
func IsNotNull(v any) bool { return objimpl.IsNotNull(v) }

// IsEmpty reports whether v is nil or an empty string, array, slice, map, or channel.
func IsEmpty(v any) bool { return objimpl.IsEmpty(v) }

// IsNotEmpty reports whether v is not empty.
func IsNotEmpty(v any) bool { return objimpl.IsNotEmpty(v) }

// DefaultIfNil returns defaultValue when object is nil.
func DefaultIfNil[T any](object *T, defaultValue T) T {
	return objimpl.DefaultIfNil(object, defaultValue)
}

// DefaultIfNilFunc returns a supplier value when object is nil.
func DefaultIfNilFunc[T any](object *T, supplier func() T) T {
	return objimpl.DefaultIfNilFunc(object, supplier)
}

// DefaultIfNilApply returns defaultValue when source is nil; otherwise it maps source.
func DefaultIfNilApply[T any, R any](source *T, handle func(T) R, defaultValue R) R {
	return objimpl.DefaultIfNilApply(source, handle, defaultValue)
}

// Apply maps source when it is not nil; otherwise it returns the zero value.
func Apply[T any, R any](source *T, handle func(T) R) R { return objimpl.Apply(source, handle) }

// Accept calls consumer when source is not nil.
func Accept[T any](source *T, consumer func(T)) { objimpl.Accept(source, consumer) }

// DefaultIfEmpty returns defaultValue when s is empty.
func DefaultIfEmpty(s, defaultValue string) string { return objimpl.DefaultIfEmpty(s, defaultValue) }

// DefaultIfEmptyFunc returns a supplier value when s is empty.
func DefaultIfEmptyFunc(s string, supplier func() string) string {
	return objimpl.DefaultIfEmptyFunc(s, supplier)
}

// DefaultIfEmptyApply returns defaultValue when s is empty; otherwise it maps s.
func DefaultIfEmptyApply[T any](s string, handle func(string) T, defaultValue T) T {
	return objimpl.DefaultIfEmptyApply(s, handle, defaultValue)
}

// DefaultIfBlank returns defaultValue when s is empty or contains only whitespace.
// String-specific blank checking is available via vstr.DefaultIfBlank.
func DefaultIfBlank(s, defaultValue string) string { return objimpl.DefaultIfBlank(s, defaultValue) }

// DefaultIfBlankFunc returns a supplier value when s is blank.
func DefaultIfBlankFunc(s string, supplier func() string) string {
	return objimpl.DefaultIfBlankFunc(s, supplier)
}

// DefaultIfBlankApply returns defaultValue when s is blank; otherwise it maps s.
func DefaultIfBlankApply[T any](s string, handle func(string) T, defaultValue T) T {
	return objimpl.DefaultIfBlankApply(s, handle, defaultValue)
}

// Clone creates a deep copy through gob serialization.
func Clone[T any](src T) (T, error) { return objimpl.Clone(src) }

// CloneIfPossible returns a cloned value when cloning succeeds, otherwise src.
func CloneIfPossible[T any](src T) T { return objimpl.CloneIfPossible(src) }

// CloneByStream creates a deep copy through gob serialization.
func CloneByStream[T any](src T) (T, error) { return objimpl.CloneByStream(src) }

// Serialize encodes obj with gob.
func Serialize[T any](obj T) ([]byte, error) { return objimpl.Serialize(obj) }

// Deserialize decodes gob data into out, which must be a pointer.
func Deserialize(data []byte, out any) error { return objimpl.Deserialize(data, out) }

// IsBasicType reports whether object is a built-in scalar type or string.
func IsBasicType(object any) bool { return objimpl.IsBasicType(object) }

// IsValidIfNumber reports false for NaN or infinite float values. Non-number values are valid.
func IsValidIfNumber(object any) bool { return objimpl.IsValidIfNumber(object) }

// Compare compares two ordered values. Nil pointers are ordered after non-nil values by default.
func Compare[T Ordered](a, b *T) int { return objimpl.Compare(a, b) }

// CompareNull compares two ordered values and controls nil ordering.
func CompareNull[T Ordered](a, b *T, nilGreater bool) int {
	return objimpl.CompareNull(a, b, nilGreater)
}

// TypeOf returns the reflection type of object, or nil for nil values.
func TypeOf(object any) reflect.Type { return objimpl.TypeOf(object) }

// TypeName returns the full type name of object.
func TypeName(object any) string { return objimpl.TypeName(object) }

// ToString converts object to a string. Nil values become "null".
func ToString(object any) string { return objimpl.ToString(object) }

// EmptyCount counts nil or empty values.
func EmptyCount(values ...any) int { return objimpl.EmptyCount(values...) }

// HasNil reports whether any value is nil.
func HasNil(values ...any) bool { return objimpl.HasNil(values...) }

// HasNull is an alias of HasNil.
func HasNull(values ...any) bool { return objimpl.HasNull(values...) }

// HasEmpty reports whether any value is nil or empty.
func HasEmpty(values ...any) bool { return objimpl.HasEmpty(values...) }

// IsAllEmpty reports whether all values are nil or empty.
func IsAllEmpty(values ...any) bool { return objimpl.IsAllEmpty(values...) }

// IsAllNotEmpty reports whether all values are not empty.
func IsAllNotEmpty(values ...any) bool { return objimpl.IsAllNotEmpty(values...) }
