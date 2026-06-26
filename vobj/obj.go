package vobj

import (
	"io"
	"reflect"

	objimpl "github.com/imajinyun/knifer-go/internal/obj"
)

// Ordered is the set of built-in ordered value types.
type Ordered interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr |
		~float32 | ~float64 | ~string
}

type (
	// Encoder is the serialization encoder contract used by object helpers.
	Encoder = objimpl.Encoder
	// Decoder is the serialization decoder contract used by object helpers.
	Decoder = objimpl.Decoder
	// CodecOption customizes object serialization helpers per call.
	CodecOption = objimpl.CodecOption
)

// WithEncoderFactory sets the encoder factory used by SerializeWithOptions and CloneWithOptions.
func WithEncoderFactory(factory func(io.Writer) Encoder) CodecOption {
	return objimpl.WithEncoderFactory(factory)
}

// WithDecoderFactory sets the decoder factory used by DeserializeWithOptions and CloneWithOptions.
func WithDecoderFactory(factory func(io.Reader) Decoder) CodecOption {
	return objimpl.WithDecoderFactory(factory)
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
// For reflection-centric workflows, vref.IsNil is the canonical source.
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

// Clone creates a deep copy through gob serialization.
func Clone[T any](src T) (T, error) { return objimpl.Clone(src) }

// CloneWithOptions creates a deep copy using per-call codec options.
func CloneWithOptions[T any](src T, opts ...CodecOption) (T, error) {
	return objimpl.CloneWithOptions(src, opts...)
}

// CloneIfPossible returns a cloned value when cloning succeeds, otherwise src.
func CloneIfPossible[T any](src T) T { return objimpl.CloneIfPossible(src) }

// CloneIfPossibleWithOptions returns a cloned value using per-call codec options when cloning succeeds, otherwise src.
func CloneIfPossibleWithOptions[T any](src T, opts ...CodecOption) T {
	return objimpl.CloneIfPossibleWithOptions(src, opts...)
}

// CloneByStream creates a deep copy through gob serialization.
func CloneByStream[T any](src T) (T, error) { return objimpl.CloneByStream(src) }

// CloneByStreamWithOptions creates a deep copy using per-call codec options.
func CloneByStreamWithOptions[T any](src T, opts ...CodecOption) (T, error) {
	return objimpl.CloneByStreamWithOptions(src, opts...)
}

// Serialize encodes obj with gob.
func Serialize[T any](obj T) ([]byte, error) { return objimpl.Serialize(obj) }

// SerializeWithOptions encodes obj using per-call codec options.
func SerializeWithOptions[T any](obj T, opts ...CodecOption) ([]byte, error) {
	return objimpl.SerializeWithOptions(obj, opts...)
}

// SerializeOrNil encodes obj with gob and returns nil when encoding fails.
func SerializeOrNil[T any](obj T) []byte { return objimpl.SerializeOrNil(obj) }

// SerializeOrNilWithOptions encodes obj using per-call codec options and returns nil when encoding fails.
func SerializeOrNilWithOptions[T any](obj T, opts ...CodecOption) []byte {
	return objimpl.SerializeOrNilWithOptions(obj, opts...)
}

// Deserialize decodes gob data into out, which must be a pointer.
//
// When acceptedTypes is not empty, the decoded object graph must contain only
// built-in container/scalar types plus values assignable to one of the accepted
// types. Accepted entries may be concrete values, pointers, or reflect.Type.
func Deserialize(data []byte, out any, acceptedTypes ...any) error {
	return objimpl.Deserialize(data, out, acceptedTypes...)
}

// DeserializeWithOptions decodes data using per-call codec options.
func DeserializeWithOptions(data []byte, out any, acceptedTypes []any, opts ...CodecOption) error {
	return objimpl.DeserializeWithOptions(data, out, acceptedTypes, opts...)
}

// DeserializeTo decodes gob data into a new value.
func DeserializeTo[T any](data []byte, acceptedTypes ...any) (T, error) {
	return objimpl.DeserializeTo[T](data, acceptedTypes...)
}

// DeserializeToWithOptions decodes data into a new value using per-call codec options.
func DeserializeToWithOptions[T any](data []byte, acceptedTypes []any, opts ...CodecOption) (T, error) {
	return objimpl.DeserializeToWithOptions[T](data, acceptedTypes, opts...)
}

// MustDeserialize decodes gob data into a new value and panics on failure.
func MustDeserialize[T any](data []byte, acceptedTypes ...any) T {
	return objimpl.MustDeserialize[T](data, acceptedTypes...)
}

// Register records a concrete type for gob interface encoding.
func Register(value any) { objimpl.Register(value) }

// RegisterName records a concrete type with a custom gob name.
func RegisterName(name string, value any) { objimpl.RegisterName(name, value) }

// ValidateAcceptedTypes checks whether value only contains built-in safe types
// plus values assignable to one of acceptedTypes.
func ValidateAcceptedTypes(value any, acceptedTypes ...any) error {
	return objimpl.ValidateAcceptedTypes(value, acceptedTypes...)
}

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
