package vser

import serializeimpl "github.com/imajinyun/go-knifer/internal/serialize"

// Clone creates a deep copy through gob serialization.
func Clone[T any](src T) (T, error) { return serializeimpl.Clone(src) }

// CloneIfPossible returns a cloned value when cloning succeeds, otherwise src.
func CloneIfPossible[T any](src T) T { return serializeimpl.CloneIfPossible(src) }

// CloneByStream creates a deep copy through gob serialization.
func CloneByStream[T any](src T) (T, error) { return serializeimpl.CloneByStream(src) }

// Serialize encodes obj with gob.
func Serialize[T any](obj T) ([]byte, error) { return serializeimpl.Serialize(obj) }

// SerializeOrNil encodes obj with gob and returns nil when encoding fails.
func SerializeOrNil[T any](obj T) []byte { return serializeimpl.SerializeOrNil(obj) }

// Deserialize decodes gob data into out, which must be a pointer.
//
// When acceptedTypes is not empty, the decoded object graph must contain only
// built-in container/scalar types plus values assignable to one of the accepted
// types. Accepted entries may be concrete values, pointers, or reflect.Type.
func Deserialize(data []byte, out any, acceptedTypes ...any) error {
	return serializeimpl.Deserialize(data, out, acceptedTypes...)
}

// DeserializeTo decodes gob data into a new value.
func DeserializeTo[T any](data []byte, acceptedTypes ...any) (T, error) {
	return serializeimpl.DeserializeTo[T](data, acceptedTypes...)
}

// MustDeserialize decodes gob data into a new value and panics on failure.
func MustDeserialize[T any](data []byte, acceptedTypes ...any) T {
	return serializeimpl.MustDeserialize[T](data, acceptedTypes...)
}

// Register records a concrete type for gob interface encoding.
func Register(value any) { serializeimpl.Register(value) }

// RegisterName records a concrete type with a custom gob name.
func RegisterName(name string, value any) { serializeimpl.RegisterName(name, value) }

// ValidateAcceptedTypes checks whether value only contains built-in safe types
// plus values assignable to one of acceptedTypes.
func ValidateAcceptedTypes(value any, acceptedTypes ...any) error {
	return serializeimpl.ValidateAcceptedTypes(value, acceptedTypes...)
}
