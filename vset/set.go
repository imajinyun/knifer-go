package vset

import (
	"fmt"

	setimpl "github.com/imajinyun/knifer-go/internal/sets"
)

// JSONOption customizes explicit Set JSON helpers per call.
type JSONOption = setimpl.JSONOption

// WithSetMarshalFunc sets the marshal provider used by MarshalJSONWithOptions.
func WithSetMarshalFunc(marshal func(any) ([]byte, error)) JSONOption {
	return setimpl.WithSetMarshalFunc(marshal)
}

// WithSetUnmarshalFunc sets the unmarshal provider used by UnmarshalJSONWithOptions.
func WithSetUnmarshalFunc(unmarshal func([]byte, any) error) JSONOption {
	return setimpl.WithSetUnmarshalFunc(unmarshal)
}

// Set is a generic set for comparable values.
type Set[T comparable] setimpl.Set[T]

// Int is a set of int values.
type Int = setimpl.Int

// Int32 is a set of int32 values.
type Int32 = setimpl.Int32

// Int64 is a set of int64 values.
type Int64 = setimpl.Int64

// Uint is a set of uint values.
type Uint = setimpl.Uint

// Uint32 is a set of uint32 values.
type Uint32 = setimpl.Uint32

// Uint64 is a set of uint64 values.
type Uint64 = setimpl.Uint64

// String is a set of string values.
type String = setimpl.String

// NewInt creates an int set.
func NewInt(items ...int) Int { return setimpl.NewInt(items...) }

// NewInt32 creates an int32 set.
func NewInt32(items ...int32) Int32 { return setimpl.NewInt32(items...) }

// NewInt64 creates an int64 set.
func NewInt64(items ...int64) Int64 { return setimpl.NewInt64(items...) }

// NewUint creates a uint set.
func NewUint(items ...uint) Uint { return setimpl.NewUint(items...) }

// NewUint32 creates a uint32 set.
func NewUint32(items ...uint32) Uint32 { return setimpl.NewUint32(items...) }

// NewUint64 creates a uint64 set.
func NewUint64(items ...uint64) Uint64 { return setimpl.NewUint64(items...) }

// NewString creates a string set.
func NewString(items ...string) String { return setimpl.NewString(items...) }

// New creates a generic set.
func New[T comparable](items ...T) Set[T] { return Set[T](setimpl.New(items...)) }

// Add inserts items into the set.
func (s Set[T]) Add(items ...T) { setimpl.Set[T](s).Add(items...) }

// Remove deletes items from the set.
func (s Set[T]) Remove(items ...T) { setimpl.Set[T](s).Remove(items...) }

// Contains reports whether item exists in the set.
func (s Set[T]) Contains(item T) bool { return setimpl.Set[T](s).Contains(item) }

// Sub returns the set difference s - other.
func (s Set[T]) Sub(other Set[T]) Set[T] {
	return Set[T](setimpl.Set[T](s).Sub(setimpl.Set[T](other)))
}

// Union returns a set containing all values from s and other.
func (s Set[T]) Union(other Set[T]) Set[T] {
	return Set[T](setimpl.Set[T](s).Union(setimpl.Set[T](other)))
}

// Intersect returns a set containing values present in both sets.
func (s Set[T]) Intersect(other Set[T]) Set[T] {
	return Set[T](setimpl.Set[T](s).Intersect(setimpl.Set[T](other)))
}

// Members returns all values in the set. The order is intentionally undefined.
func (s Set[T]) Members() []T { return setimpl.Set[T](s).Members() }

// Equal reports whether both sets contain exactly the same values.
func (s Set[T]) Equal(other Set[T]) bool { return setimpl.Set[T](s).Equal(setimpl.Set[T](other)) }

// String returns a human-readable representation of the set.
func (s Set[T]) String() string { return fmt.Sprintf("set%v", s.Members()) }

// MarshalJSON encodes the set as a JSON array.
func (s Set[T]) MarshalJSON() ([]byte, error) { return s.MarshalJSONWithOptions() }

// MarshalJSONWithOptions encodes the set as a JSON array with options.
func (s Set[T]) MarshalJSONWithOptions(opts ...JSONOption) ([]byte, error) {
	return setimpl.Set[T](s).MarshalJSONWithOptions(opts...)
}

// UnmarshalJSON decodes a JSON array into the set.
func (s *Set[T]) UnmarshalJSON(data []byte) error {
	return s.UnmarshalJSONWithOptions(data)
}

// UnmarshalJSONWithOptions decodes a JSON array into the set with options.
func (s *Set[T]) UnmarshalJSONWithOptions(data []byte, opts ...JSONOption) error {
	inner := setimpl.Set[T](*s)
	if err := (&inner).UnmarshalJSONWithOptions(data, opts...); err != nil {
		return err
	}
	*s = Set[T](inner)
	return nil
}

// MarshalYAML encodes the set as a YAML sequence.
func (s Set[T]) MarshalYAML() (any, error) { return s.Members(), nil }

// UnmarshalYAML decodes a YAML sequence into the set.
func (s *Set[T]) UnmarshalYAML(unmarshal func(any) error) error {
	var list []T
	if err := unmarshal(&list); err != nil {
		return err
	}
	*s = New(list...)
	return nil
}
