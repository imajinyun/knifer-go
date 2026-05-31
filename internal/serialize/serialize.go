package serialize

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"reflect"
)

// Clone creates a deep copy through gob serialization.
func Clone[T any](src T) (T, error) {
	data, err := Serialize(src)
	if err != nil {
		var zero T
		return zero, err
	}
	return DeserializeTo[T](data)
}

// CloneIfPossible returns a cloned value when cloning succeeds, otherwise src.
func CloneIfPossible[T any](src T) T {
	clone, err := Clone(src)
	if err != nil {
		return src
	}
	return clone
}

// CloneByStream creates a deep copy through gob serialization.
func CloneByStream[T any](src T) (T, error) { return Clone(src) }

// Serialize encodes obj with gob.
func Serialize[T any](obj T) ([]byte, error) {
	var buf bytes.Buffer
	err := gob.NewEncoder(&buf).Encode(obj)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// SerializeOrNil encodes obj with gob and returns nil when encoding fails.
func SerializeOrNil[T any](obj T) []byte {
	data, err := Serialize(obj)
	if err != nil {
		return nil
	}
	return data
}

// Deserialize decodes gob data into out, which must be a pointer.
//
// When acceptedTypes is not empty, the decoded object graph must contain only
// built-in container/scalar types plus values assignable to one of the accepted
// types. Accepted entries may be concrete values, pointers, or reflect.Type.
func Deserialize(data []byte, out any, acceptedTypes ...any) error {
	if err := gob.NewDecoder(bytes.NewReader(data)).Decode(out); err != nil {
		return err
	}
	if len(acceptedTypes) == 0 {
		return nil
	}
	return ValidateAcceptedTypes(out, acceptedTypes...)
}

// DeserializeTo decodes gob data into a new value.
func DeserializeTo[T any](data []byte, acceptedTypes ...any) (T, error) {
	var out T
	if err := Deserialize(data, &out, acceptedTypes...); err != nil {
		var zero T
		return zero, err
	}
	return out, nil
}

// MustDeserialize decodes gob data into a new value and panics on failure.
func MustDeserialize[T any](data []byte, acceptedTypes ...any) T {
	out, err := DeserializeTo[T](data, acceptedTypes...)
	if err != nil {
		panic(err)
	}
	return out
}

// Register records a concrete type for gob interface encoding.
func Register(value any) { gob.Register(value) }

// RegisterName records a concrete type with a custom gob name.
func RegisterName(name string, value any) { gob.RegisterName(name, value) }

// ValidateAcceptedTypes checks whether value only contains built-in safe types
// plus values assignable to one of acceptedTypes.
func ValidateAcceptedTypes(value any, acceptedTypes ...any) error {
	if len(acceptedTypes) == 0 {
		return nil
	}
	allowed := make([]reflect.Type, 0, len(acceptedTypes))
	for _, accepted := range acceptedTypes {
		if accepted == nil {
			continue
		}
		if t, ok := accepted.(reflect.Type); ok {
			allowed = append(allowed, t)
			continue
		}
		allowed = append(allowed, reflect.TypeOf(accepted))
	}
	return validateValue(reflect.ValueOf(value), allowed, map[visit]bool{})
}

type visit struct {
	typ reflect.Type
	ptr uintptr
}

func validateValue(v reflect.Value, allowed []reflect.Type, seen map[visit]bool) error {
	if !v.IsValid() {
		return nil
	}
	for v.Kind() == reflect.Interface || v.Kind() == reflect.Pointer {
		if v.IsNil() {
			return nil
		}
		if v.Kind() == reflect.Pointer {
			key := visit{typ: v.Type(), ptr: v.Pointer()}
			if seen[key] {
				return nil
			}
			seen[key] = true
		}
		v = v.Elem()
	}
	t := v.Type()
	if isAllowedType(t, allowed) || isBuiltInAllowedKind(v.Kind()) {
		return validateChildren(v, allowed, seen)
	}
	return fmt.Errorf("serialize: decoded type %s is not accepted", t)
}

func validateChildren(v reflect.Value, allowed []reflect.Type, seen map[visit]bool) error {
	switch v.Kind() {
	case reflect.Array, reflect.Slice:
		for i := 0; i < v.Len(); i++ {
			if err := validateValue(v.Index(i), allowed, seen); err != nil {
				return err
			}
		}
	case reflect.Map:
		iter := v.MapRange()
		for iter.Next() {
			if err := validateValue(iter.Key(), allowed, seen); err != nil {
				return err
			}
			if err := validateValue(iter.Value(), allowed, seen); err != nil {
				return err
			}
		}
	case reflect.Struct:
		for i := 0; i < v.NumField(); i++ {
			field := v.Field(i)
			if field.CanInterface() {
				if err := validateValue(field, allowed, seen); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func isAllowedType(t reflect.Type, allowed []reflect.Type) bool {
	for _, a := range allowed {
		if t.AssignableTo(a) || reflect.PointerTo(t).AssignableTo(a) || t.AssignableTo(indirectType(a)) {
			return true
		}
	}
	return false
}

func indirectType(t reflect.Type) reflect.Type {
	for t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	return t
}

func isBuiltInAllowedKind(kind reflect.Kind) bool {
	switch kind {
	case reflect.Bool,
		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr,
		reflect.Float32, reflect.Float64,
		reflect.Complex64, reflect.Complex128,
		reflect.String,
		reflect.Array, reflect.Slice, reflect.Map:
		return true
	default:
		return false
	}
}
