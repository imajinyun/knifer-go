package sets

import (
	"encoding/json"
	"fmt"
)

type jsonConfig struct {
	marshal   func(any) ([]byte, error)
	unmarshal func([]byte, any) error
}

// JSONOption customizes explicit Set JSON helpers per call.
type JSONOption func(*jsonConfig)

// WithSetMarshalFunc sets the marshal provider used by MarshalJSONWithOptions.
func WithSetMarshalFunc(marshal func(any) ([]byte, error)) JSONOption {
	return func(c *jsonConfig) {
		if marshal != nil {
			c.marshal = marshal
		}
	}
}

// WithSetUnmarshalFunc sets the unmarshal provider used by UnmarshalJSONWithOptions.
func WithSetUnmarshalFunc(unmarshal func([]byte, any) error) JSONOption {
	return func(c *jsonConfig) {
		if unmarshal != nil {
			c.unmarshal = unmarshal
		}
	}
}

func applyJSONOptions(opts []JSONOption) jsonConfig {
	cfg := jsonConfig{marshal: json.Marshal, unmarshal: json.Unmarshal}
	for _, opt := range opts {
		if opt != nil {
			opt(&cfg)
		}
	}
	if cfg.marshal == nil {
		cfg.marshal = json.Marshal
	}
	if cfg.unmarshal == nil {
		cfg.unmarshal = json.Unmarshal
	}
	return cfg
}

// Set is a generic hash set for comparable values.
type Set[T comparable] map[T]struct{}

// New creates a generic set with the given items.
func New[T comparable](items ...T) Set[T] {
	s := make(Set[T], len(items))
	s.Add(items...)
	return s
}

// Add inserts items into the set.
func (s Set[T]) Add(items ...T) {
	for _, item := range items {
		s[item] = struct{}{}
	}
}

// Remove deletes items from the set.
func (s Set[T]) Remove(items ...T) {
	for _, item := range items {
		delete(s, item)
	}
}

// Contains reports whether item exists in the set.
func (s Set[T]) Contains(item T) bool {
	_, ok := s[item]
	return ok
}

// Sub returns the set difference s - other.
func (s Set[T]) Sub(other Set[T]) Set[T] {
	out := make(Set[T], len(s))
	for item := range s {
		if !other.Contains(item) {
			out[item] = struct{}{}
		}
	}
	return out
}

// Union returns a set containing all values from s and other.
func (s Set[T]) Union(other Set[T]) Set[T] {
	out := make(Set[T], len(s)+len(other))
	for item := range s {
		out[item] = struct{}{}
	}
	for item := range other {
		out[item] = struct{}{}
	}
	return out
}

// Intersect returns a set containing values present in both sets.
func (s Set[T]) Intersect(other Set[T]) Set[T] {
	if len(other) < len(s) {
		return other.Intersect(s)
	}
	out := make(Set[T])
	for item := range s {
		if other.Contains(item) {
			out[item] = struct{}{}
		}
	}
	return out
}

// Members returns all values in the set. The order is intentionally undefined.
func (s Set[T]) Members() []T {
	items := make([]T, 0, len(s))
	for item := range s {
		items = append(items, item)
	}
	return items
}

// Equal reports whether both sets contain exactly the same values.
func (s Set[T]) Equal(other Set[T]) bool {
	if len(s) != len(other) {
		return false
	}
	for item := range s {
		if !other.Contains(item) {
			return false
		}
	}
	return true
}

// String returns a human-readable representation of the set.
func (s Set[T]) String() string { return fmt.Sprintf("set%v", s.Members()) }

// MarshalJSON encodes the set as a JSON array.
func (s Set[T]) MarshalJSON() ([]byte, error) { return s.MarshalJSONWithOptions() }

// MarshalJSONWithOptions encodes the set as a JSON array with options.
func (s Set[T]) MarshalJSONWithOptions(opts ...JSONOption) ([]byte, error) {
	return applyJSONOptions(opts).marshal(s.Members())
}

// UnmarshalJSON decodes a JSON array into the set.
func (s *Set[T]) UnmarshalJSON(data []byte) error {
	return s.UnmarshalJSONWithOptions(data)
}

// UnmarshalJSONWithOptions decodes a JSON array into the set with options.
func (s *Set[T]) UnmarshalJSONWithOptions(data []byte, opts ...JSONOption) error {
	var list []T
	if err := applyJSONOptions(opts).unmarshal(data, &list); err != nil {
		return err
	}
	*s = New(list...)
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
