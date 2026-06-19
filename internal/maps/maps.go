// Package maps provides common helpers for Go map collections.
// All functions are pure: input maps are never mutated, return values are never nil
// unless explicitly documented.
package maps

import (
	"cmp"
	"fmt"
	"iter"
	stdmaps "maps"
	"slices"
)

// ---------------------------------------------------------------------------
// Construction
// ---------------------------------------------------------------------------

// New returns an empty, non-nil map.
func New[K comparable, V any]() map[K]V { return make(map[K]V) }

// NewWithCap returns an empty, non-nil map pre-sized to hint.
func NewWithCap[K comparable, V any](hint int) map[K]V {
	if hint < 0 {
		hint = 0
	}
	return make(map[K]V, hint)
}

// Pair stores one typed key-value pair for FromPairs.
type Pair[K comparable, V any] struct {
	Key   K
	Value V
}

// Of builds a map from alternating key-value pairs.
// Panics if len(kvs) is odd. Later duplicate keys override earlier ones.
//
//	Of[string, int]("a", 1, "b", 2) // map[string]int{"a": 1, "b": 2}
func Of[K comparable, V any](kvs ...any) map[K]V {
	if len(kvs)%2 != 0 {
		panic("maps.Of: odd number of arguments")
	}
	out := make(map[K]V, len(kvs)/2)
	for i := 0; i < len(kvs); i += 2 {
		out[kvs[i].(K)] = kvs[i+1].(V)
	}
	return out
}

// OfE builds a map from alternating key-value pairs and returns errors instead of panicking.
// Later duplicate keys override earlier ones.
func OfE[K comparable, V any](kvs ...any) (map[K]V, error) {
	if len(kvs)%2 != 0 {
		return nil, fmt.Errorf("maps.OfE: odd number of arguments")
	}
	out := make(map[K]V, len(kvs)/2)
	for i := 0; i < len(kvs); i += 2 {
		key, ok := kvs[i].(K)
		if !ok {
			return nil, fmt.Errorf("maps.OfE: argument %d has type %T, want key type", i, kvs[i])
		}
		value, ok := kvs[i+1].(V)
		if !ok {
			return nil, fmt.Errorf("maps.OfE: argument %d has type %T, want value type", i+1, kvs[i+1])
		}
		out[key] = value
	}
	return out, nil
}

// FromPairs builds a map from typed key-value pairs.
// Later duplicate keys override earlier ones.
func FromPairs[K comparable, V any](pairs ...Pair[K, V]) map[K]V {
	out := make(map[K]V, len(pairs))
	for _, pair := range pairs {
		out[pair.Key] = pair.Value
	}
	return out
}

// FromEntries builds a map from typed key-value entries.
// Later duplicate keys override earlier ones.
func FromEntries[K comparable, V any](entries []Pair[K, V]) map[K]V {
	out := make(map[K]V, len(entries))
	for _, entry := range entries {
		out[entry.Key] = entry.Value
	}
	return out
}

// OrEmpty returns m if non-nil, otherwise an empty map.
func OrEmpty[K comparable, V any](m map[K]V) map[K]V {
	if m == nil {
		return make(map[K]V)
	}
	return m
}

// ---------------------------------------------------------------------------
// Predicates
// ---------------------------------------------------------------------------

// IsEmpty reports whether the map is empty.
func IsEmpty[K comparable, V any](m map[K]V) bool { return len(m) == 0 }

// IsNotEmpty reports whether the map is not empty.
func IsNotEmpty[K comparable, V any](m map[K]V) bool { return len(m) > 0 }

// ContainsKey reports whether key exists in m.
func ContainsKey[K comparable, V any](m map[K]V, key K) bool {
	_, ok := m[key]
	return ok
}

// ContainsValue reports whether any entry of m has the given value.
// V must be comparable.
func ContainsValue[K, V comparable](m map[K]V, value V) bool {
	return slices.Contains(slices.Collect(stdmaps.Values(m)), value)
}

// Some reports whether at least one entry satisfies the predicate.
func Some[K comparable, V any](m map[K]V, pred func(K, V) bool) bool {
	return slices.ContainsFunc(Entries(m), func(entry Pair[K, V]) bool {
		return pred(entry.Key, entry.Value)
	})
}

// Every reports whether every entry satisfies the predicate. Empty maps return true.
func Every[K comparable, V any](m map[K]V, pred func(K, V) bool) bool {
	return !Some(m, func(k K, v V) bool { return !pred(k, v) })
}

// ---------------------------------------------------------------------------
// Lookup
// ---------------------------------------------------------------------------

// Get returns the value for key, or the zero value of V when absent.
func Get[K comparable, V any](m map[K]V, key K) V {
	return m[key]
}

// GetOr returns the value for key, or fallback when absent.
func GetOr[K comparable, V any](m map[K]V, key K, fallback V) V {
	if v, ok := m[key]; ok {
		return v
	}
	return fallback
}

// GetAny returns the first present value among keys, or zero value when none exist.
func GetAny[K comparable, V any](m map[K]V, keys ...K) (V, bool) {
	for _, k := range keys {
		if v, ok := m[k]; ok {
			return v, true
		}
	}
	var zero V
	return zero, false
}

// Find returns the first (key, value) satisfying the predicate, plus a found flag.
// Iteration order follows Go map semantics and is not stable.
func Find[K comparable, V any](m map[K]V, pred func(K, V) bool) (K, V, bool) {
	for k, v := range m {
		if pred(k, v) {
			return k, v, true
		}
	}
	var (
		zk K
		zv V
	)
	return zk, zv, false
}

// FindKey returns the first key whose value satisfies the predicate.
func FindKey[K comparable, V any](m map[K]V, pred func(V) bool) (K, bool) {
	for k, v := range m {
		if pred(v) {
			return k, true
		}
	}
	var zk K
	return zk, false
}

// ---------------------------------------------------------------------------
// Collection views
// ---------------------------------------------------------------------------

// Keys returns all keys. Order follows Go map iteration and is not stable.
func Keys[K comparable, V any](m map[K]V) []K {
	return slices.Collect(stdmaps.Keys(m))
}

// Values returns all values. Order follows Go map iteration and is not stable.
func Values[K comparable, V any](m map[K]V) []V {
	return slices.Collect(stdmaps.Values(m))
}

// SortedKeys returns all keys sorted in ascending order. K must be ordered.
func SortedKeys[K cmp.Ordered, V any](m map[K]V) []K {
	return slices.Sorted(stdmaps.Keys(m))
}

// SortedKeysFunc returns all keys sorted by the supplied less function.
func SortedKeysFunc[K comparable, V any](m map[K]V, less func(a, b K) bool) []K {
	return slices.SortedFunc(stdmaps.Keys(m), func(a, b K) int {
		switch {
		case less(a, b):
			return -1
		case less(b, a):
			return 1
		default:
			return 0
		}
	})
}

// SortedValues returns values sorted by their keys ascending. K must be ordered.
func SortedValues[K cmp.Ordered, V any](m map[K]V) []V {
	ks := SortedKeys(m)
	vs := make([]V, 0, len(ks))
	for _, k := range ks {
		vs = append(vs, m[k])
	}
	return vs
}

// KeysOf returns all keys whose value equals target. V must be comparable.
func KeysOf[K, V comparable](m map[K]V, target V) []K {
	out := make([]K, 0)
	for k, v := range m {
		if v == target {
			out = append(out, k)
		}
	}
	return out
}

// Entries returns all key-value pairs. Order follows Go map iteration and is not stable.
func Entries[K comparable, V any](m map[K]V) []Pair[K, V] {
	out := make([]Pair[K, V], 0, len(m))
	for k, v := range m {
		out = append(out, Pair[K, V]{Key: k, Value: v})
	}
	return out
}

// Iter returns an iterator over map key-value pairs.
// Iteration order follows Go map semantics and is not stable.
func Iter[K comparable, V any](m map[K]V) iter.Seq2[K, V] { return stdmaps.All(m) }

// IterKeys returns an iterator over map keys.
// Iteration order follows Go map semantics and is not stable.
func IterKeys[K comparable, V any](m map[K]V) iter.Seq[K] { return stdmaps.Keys(m) }

// IterValues returns an iterator over map values.
// Iteration order follows Go map semantics and is not stable.
func IterValues[K comparable, V any](m map[K]V) iter.Seq[V] { return stdmaps.Values(m) }

// ---------------------------------------------------------------------------
// Transformation
// ---------------------------------------------------------------------------

// Map transforms each (k, v) into a new pair (k2, v2).
// If transform yields duplicate keys, later ones win.
func Map[K1, K2 comparable, V1, V2 any](m map[K1]V1, transform func(K1, V1) (K2, V2)) map[K2]V2 {
	out := make(map[K2]V2, len(m))
	for k, v := range m {
		k2, v2 := transform(k, v)
		out[k2] = v2
	}
	return out
}

// MapErr transforms each (k, v) into a new pair and stops on the first error.
// If transform yields duplicate keys, later successful entries win.
func MapErr[K1, K2 comparable, V1, V2 any](m map[K1]V1, transform func(K1, V1) (K2, V2, error)) (map[K2]V2, error) {
	out := make(map[K2]V2, len(m))
	for k, v := range m {
		k2, v2, err := transform(k, v)
		if err != nil {
			return out, err
		}
		out[k2] = v2
	}
	return out, nil
}

// MapKeys transforms each key, preserving values.
func MapKeys[K1, K2 comparable, V any](m map[K1]V, transform func(K1, V) K2) map[K2]V {
	out := make(map[K2]V, len(m))
	for k, v := range m {
		out[transform(k, v)] = v
	}
	return out
}

// MapKeysErr transforms each key, preserving values, and stops on the first error.
func MapKeysErr[K1, K2 comparable, V any](m map[K1]V, transform func(K1, V) (K2, error)) (map[K2]V, error) {
	out := make(map[K2]V, len(m))
	for k, v := range m {
		k2, err := transform(k, v)
		if err != nil {
			return out, err
		}
		out[k2] = v
	}
	return out, nil
}

// MapValues transforms each value, preserving keys.
func MapValues[K comparable, V1, V2 any](m map[K]V1, transform func(K, V1) V2) map[K]V2 {
	out := make(map[K]V2, len(m))
	for k, v := range m {
		out[k] = transform(k, v)
	}
	return out
}

// MapValuesErr transforms each value, preserving keys, and stops on the first error.
func MapValuesErr[K comparable, V1, V2 any](m map[K]V1, transform func(K, V1) (V2, error)) (map[K]V2, error) {
	out := make(map[K]V2, len(m))
	for k, v := range m {
		v2, err := transform(k, v)
		if err != nil {
			return out, err
		}
		out[k] = v2
	}
	return out, nil
}

// ToSlice transforms each map entry into a slice element.
// Order follows Go map iteration and is not stable.
func ToSlice[K comparable, V, R any](m map[K]V, transform func(K, V) R) []R {
	out := make([]R, 0, len(m))
	for k, v := range m {
		out = append(out, transform(k, v))
	}
	return out
}

// Filter returns entries satisfying the predicate.
func Filter[K comparable, V any](m map[K]V, pred func(K, V) bool) map[K]V {
	out := make(map[K]V)
	for k, v := range m {
		if pred(k, v) {
			out[k] = v
		}
	}
	return out
}

// FilterErr returns entries satisfying the predicate and stops on the first error.
func FilterErr[K comparable, V any](m map[K]V, pred func(K, V) (bool, error)) (map[K]V, error) {
	out := make(map[K]V)
	for k, v := range m {
		keep, err := pred(k, v)
		if err != nil {
			return out, err
		}
		if keep {
			out[k] = v
		}
	}
	return out, nil
}

// Reject returns entries NOT satisfying the predicate.
func Reject[K comparable, V any](m map[K]V, pred func(K, V) bool) map[K]V {
	return Filter(m, func(k K, v V) bool { return !pred(k, v) })
}

// FilterKeys keeps entries whose key satisfies the predicate.
func FilterKeys[K comparable, V any](m map[K]V, pred func(K) bool) map[K]V {
	return Filter(m, func(k K, _ V) bool { return pred(k) })
}

// FilterValues keeps entries whose value satisfies the predicate.
func FilterValues[K comparable, V any](m map[K]V, pred func(V) bool) map[K]V {
	return Filter(m, func(_ K, v V) bool { return pred(v) })
}

// PickBy returns entries satisfying the predicate.
func PickBy[K comparable, V any](m map[K]V, pred func(K, V) bool) map[K]V {
	return Filter(m, pred)
}

// OmitBy returns entries NOT satisfying the predicate.
func OmitBy[K comparable, V any](m map[K]V, pred func(K, V) bool) map[K]V {
	return Reject(m, pred)
}

// Partition splits m into two maps: matched and rest, by predicate.
func Partition[K comparable, V any](m map[K]V, pred func(K, V) bool) (matched, rest map[K]V) {
	matched = make(map[K]V)
	rest = make(map[K]V)
	for k, v := range m {
		if pred(k, v) {
			matched[k] = v
		} else {
			rest[k] = v
		}
	}
	return matched, rest
}

// ForEach invokes fn for every entry.
func ForEach[K comparable, V any](m map[K]V, fn func(K, V)) {
	for k, v := range m {
		fn(k, v)
	}
}

// ---------------------------------------------------------------------------
// Aggregation
// ---------------------------------------------------------------------------

// Reduce folds the map into a single value.
func Reduce[K comparable, V, R any](m map[K]V, initial R, fn func(acc R, k K, v V) R) R {
	acc := initial
	for k, v := range m {
		acc = fn(acc, k, v)
	}
	return acc
}

// ReduceErr folds the map into a single value and stops on the first error.
// The returned accumulator is the last successful accumulator value.
func ReduceErr[K comparable, V, R any](m map[K]V, initial R, fn func(acc R, k K, v V) (R, error)) (R, error) {
	acc := initial
	for k, v := range m {
		next, err := fn(acc, k, v)
		if err != nil {
			return acc, err
		}
		acc = next
	}
	return acc, nil
}

// GroupBy groups slice items into a map keyed by the result of keyFn.
func GroupBy[T any, K comparable](items []T, keyFn func(T) K) map[K][]T {
	out := make(map[K][]T)
	for _, item := range items {
		k := keyFn(item)
		out[k] = append(out[k], item)
	}
	return out
}

// CountBy counts slice items grouped by keyFn.
func CountBy[T any, K comparable](items []T, keyFn func(T) K) map[K]int {
	out := make(map[K]int)
	for _, item := range items {
		out[keyFn(item)]++
	}
	return out
}

// ---------------------------------------------------------------------------
// Set algebra
// ---------------------------------------------------------------------------

// Inverse swaps keys and values. V must be comparable; on duplicate values,
// later iterations override earlier keys.
func Inverse[K, V comparable](m map[K]V) map[V]K {
	out := make(map[V]K, len(m))
	for k, v := range m {
		out[v] = k
	}
	return out
}

// Invert swaps keys and values. It is an alias of Inverse for lo-style naming.
func Invert[K, V comparable](m map[K]V) map[V]K {
	return Inverse(m)
}

// Intersect returns entries whose keys appear in every input map.
// For duplicate keys, the value comes from the last map (consistent with Merge).
func Intersect[K comparable, V any](ms ...map[K]V) map[K]V {
	if len(ms) == 0 {
		return make(map[K]V)
	}
	if len(ms) == 1 {
		return Merge(ms[0])
	}
	// Pick the smallest map as the seed for fewer probes.
	seedIdx := 0
	for i, m := range ms {
		if len(m) < len(ms[seedIdx]) {
			seedIdx = i
		}
	}
	out := make(map[K]V)
seedLoop:
	for k := range ms[seedIdx] {
		var last V
		for _, m := range ms {
			v, ok := m[k]
			if !ok {
				continue seedLoop
			}
			last = v
		}
		out[k] = last
	}
	return out
}

// Diff returns entries from a whose keys are absent in any of others.
func Diff[K comparable, V any](a map[K]V, others ...map[K]V) map[K]V {
	out := make(map[K]V)
keyLoop:
	for k, v := range a {
		for _, o := range others {
			if _, ok := o[k]; ok {
				continue keyLoop
			}
		}
		out[k] = v
	}
	return out
}

// SymmetricDiff returns entries present in exactly one of the two maps.
func SymmetricDiff[K comparable, V any](a, b map[K]V) map[K]V {
	out := make(map[K]V, len(a)+len(b))
	for k, v := range a {
		if _, ok := b[k]; !ok {
			out[k] = v
		}
	}
	for k, v := range b {
		if _, ok := a[k]; !ok {
			out[k] = v
		}
	}
	return out
}

// ---------------------------------------------------------------------------
// Selection
// ---------------------------------------------------------------------------

// Pick returns a new map containing only the requested keys.
func Pick[K comparable, V any](m map[K]V, keys ...K) map[K]V {
	out := make(map[K]V, len(keys))
	for _, k := range keys {
		if v, ok := m[k]; ok {
			out[k] = v
		}
	}
	return out
}

// Omit returns a new map without the specified keys.
func Omit[K comparable, V any](m map[K]V, keys ...K) map[K]V {
	skip := make(map[K]struct{}, len(keys))
	for _, k := range keys {
		skip[k] = struct{}{}
	}
	out := make(map[K]V, len(m))
	for k, v := range m {
		if _, drop := skip[k]; drop {
			continue
		}
		out[k] = v
	}
	return out
}

// Assign merges maps into a new map. Later maps override earlier ones on duplicate keys.
func Assign[K comparable, V any](maps ...map[K]V) map[K]V {
	return Merge(maps...)
}

// ---------------------------------------------------------------------------
// Mutation helpers (operate in place; caller decides whether to clone first)
// ---------------------------------------------------------------------------

// Clear removes all entries from m in place.
func Clear[K comparable, V any](m map[K]V) {
	clear(m)
}

// Update copies all entries from src into dst, overriding existing keys.
// Returns dst for chaining. A nil dst is treated as a fresh map.
func Update[K comparable, V any](dst, src map[K]V) map[K]V {
	if dst == nil {
		dst = make(map[K]V, len(src))
	}
	stdmaps.Copy(dst, src)
	return dst
}

// Clone returns a shallow copy of m. Returns an empty map when m is nil.
func Clone[K comparable, V any](m map[K]V) map[K]V {
	if m == nil {
		return make(map[K]V)
	}
	return stdmaps.Clone(m)
}

// ---------------------------------------------------------------------------
// Comparison
// ---------------------------------------------------------------------------

// Equal reports whether two maps contain the same key-value pairs.
// V must be comparable.
func Equal[K, V comparable](a, b map[K]V) bool {
	return stdmaps.Equal(a, b)
}

// EqualFunc reports whether two maps contain the same keys and pairwise-equivalent
// values per eq.
func EqualFunc[K comparable, V1, V2 any](a map[K]V1, b map[K]V2, eq func(V1, V2) bool) bool {
	return stdmaps.EqualFunc(a, b, eq)
}
