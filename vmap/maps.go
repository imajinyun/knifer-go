package vmap

import (
	"cmp"
	"iter"

	mapsimpl "github.com/imajinyun/go-knifer/internal/maps"
)

type Pair[K comparable, V any] = mapsimpl.Pair[K, V]

// New creates an initialized empty map.
func New[K comparable, V any]() map[K]V { return mapsimpl.New[K, V]() }

// NewWithCap creates an initialized empty map with a capacity hint.
func NewWithCap[K comparable, V any](hint int) map[K]V { return mapsimpl.NewWithCap[K, V](hint) }

// Of creates a map from alternating key and value arguments and drops invalid pairs.
func Of[K comparable, V any](kvs ...any) map[K]V { return mapsimpl.Of[K, V](kvs...) }

// OfE creates a map from alternating key and value arguments and reports invalid input.
func OfE[K comparable, V any](kvs ...any) (map[K]V, error) {
	return mapsimpl.OfE[K, V](kvs...)
}

func FromPairs[K comparable, V any](pairs ...Pair[K, V]) map[K]V {
	return mapsimpl.FromPairs(pairs...)
}

func FromEntries[K comparable, V any](entries []Pair[K, V]) map[K]V {
	return mapsimpl.FromEntries(entries)
}
func OrEmpty[K comparable, V any](m map[K]V) map[K]V { return mapsimpl.OrEmpty(m) }
func IsEmpty[K comparable, V any](m map[K]V) bool    { return mapsimpl.IsEmpty(m) }
func IsNotEmpty[K comparable, V any](m map[K]V) bool { return mapsimpl.IsNotEmpty(m) }

func ContainsKey[K comparable, V any](m map[K]V, key K) bool { return mapsimpl.ContainsKey(m, key) }

func ContainsValue[K, V comparable](m map[K]V, value V) bool { return mapsimpl.ContainsValue(m, value) }

func Some[K comparable, V any](m map[K]V, pred func(K, V) bool) bool { return mapsimpl.Some(m, pred) }

func Every[K comparable, V any](m map[K]V, pred func(K, V) bool) bool { return mapsimpl.Every(m, pred) }

func Get[K comparable, V any](m map[K]V, key K) V { return mapsimpl.Get(m, key) }

func GetOr[K comparable, V any](m map[K]V, key K, fallback V) V {
	return mapsimpl.GetOr(m, key, fallback)
}

func GetAny[K comparable, V any](m map[K]V, keys ...K) (V, bool) { return mapsimpl.GetAny(m, keys...) }

func Find[K comparable, V any](m map[K]V, pred func(K, V) bool) (K, V, bool) {
	return mapsimpl.Find(m, pred)
}

func FindKey[K comparable, V any](m map[K]V, pred func(V) bool) (K, bool) {
	return mapsimpl.FindKey(m, pred)
}
func Keys[K comparable, V any](m map[K]V) []K        { return mapsimpl.Keys(m) }
func Values[K comparable, V any](m map[K]V) []V      { return mapsimpl.Values(m) }
func SortedKeys[K cmp.Ordered, V any](m map[K]V) []K { return mapsimpl.SortedKeys(m) }
func SortedKeysFunc[K comparable, V any](m map[K]V, less func(a, b K) bool) []K {
	return mapsimpl.SortedKeysFunc(m, less)
}
func SortedValues[K cmp.Ordered, V any](m map[K]V) []V { return mapsimpl.SortedValues(m) }
func KeysOf[K, V comparable](m map[K]V, target V) []K  { return mapsimpl.KeysOf(m, target) }
func Entries[K comparable, V any](m map[K]V) []Pair[K, V] {
	return mapsimpl.Entries(m)
}
func Iter[K comparable, V any](m map[K]V) iter.Seq2[K, V] { return mapsimpl.Iter(m) }
func IterKeys[K comparable, V any](m map[K]V) iter.Seq[K] { return mapsimpl.IterKeys(m) }
func IterValues[K comparable, V any](m map[K]V) iter.Seq[V] {
	return mapsimpl.IterValues(m)
}

func Map[K1, K2 comparable, V1, V2 any](m map[K1]V1, transform func(K1, V1) (K2, V2)) map[K2]V2 {
	return mapsimpl.Map(m, transform)
}

func MapKeys[K1, K2 comparable, V any](m map[K1]V, transform func(K1, V) K2) map[K2]V {
	return mapsimpl.MapKeys(m, transform)
}

func MapValues[K comparable, V1, V2 any](m map[K]V1, transform func(K, V1) V2) map[K]V2 {
	return mapsimpl.MapValues(m, transform)
}

func ToSlice[K comparable, V, R any](m map[K]V, transform func(K, V) R) []R {
	return mapsimpl.ToSlice(m, transform)
}

func Filter[K comparable, V any](m map[K]V, pred func(K, V) bool) map[K]V {
	return mapsimpl.Filter(m, pred)
}

func Reject[K comparable, V any](m map[K]V, pred func(K, V) bool) map[K]V {
	return mapsimpl.Reject(m, pred)
}

func FilterKeys[K comparable, V any](m map[K]V, pred func(K) bool) map[K]V {
	return mapsimpl.FilterKeys(m, pred)
}

func FilterValues[K comparable, V any](m map[K]V, pred func(V) bool) map[K]V {
	return mapsimpl.FilterValues(m, pred)
}

func PickBy[K comparable, V any](m map[K]V, pred func(K, V) bool) map[K]V {
	return mapsimpl.PickBy(m, pred)
}

func OmitBy[K comparable, V any](m map[K]V, pred func(K, V) bool) map[K]V {
	return mapsimpl.OmitBy(m, pred)
}

func Partition[K comparable, V any](m map[K]V, pred func(K, V) bool) (map[K]V, map[K]V) {
	return mapsimpl.Partition(m, pred)
}
func ForEach[K comparable, V any](m map[K]V, fn func(K, V)) { mapsimpl.ForEach(m, fn) }
func Reduce[K comparable, V, R any](m map[K]V, initial R, fn func(R, K, V) R) R {
	return mapsimpl.Reduce(m, initial, fn)
}

func GroupBy[T any, K comparable](items []T, keyFn func(T) K) map[K][]T {
	return mapsimpl.GroupBy(items, keyFn)
}

func CountBy[T any, K comparable](items []T, keyFn func(T) K) map[K]int {
	return mapsimpl.CountBy(items, keyFn)
}
func Inverse[K, V comparable](m map[K]V) map[V]K         { return mapsimpl.Inverse(m) }
func Invert[K, V comparable](m map[K]V) map[V]K          { return mapsimpl.Invert(m) }
func Merge[K comparable, V any](maps ...map[K]V) map[K]V { return mapsimpl.Merge(maps...) }
func Assign[K comparable, V any](maps ...map[K]V) map[K]V {
	return mapsimpl.Assign(maps...)
}

func MergeWithOverwrite[K comparable, V any](dstMap map[K]V, srcMaps ...map[K]V) {
	mapsimpl.MergeWithOverwrite(dstMap, srcMaps...)
}

func MergeWithoutOverwrite[K comparable, V any](dstMap map[K]V, srcMaps ...map[K]V) {
	mapsimpl.MergeWithoutOverwrite(dstMap, srcMaps...)
}

func MergeFunc[K comparable, V any](resolve func(old, new V) V, maps ...map[K]V) map[K]V {
	return mapsimpl.MergeFunc(resolve, maps...)
}
func Intersect[K comparable, V any](maps ...map[K]V) map[K]V { return mapsimpl.Intersect(maps...) }
func Diff[K comparable, V any](a map[K]V, others ...map[K]V) map[K]V {
	return mapsimpl.Diff(a, others...)
}

func SymmetricDiff[K comparable, V any](a, b map[K]V) map[K]V { return mapsimpl.SymmetricDiff(a, b) }
func Pick[K comparable, V any](m map[K]V, keys ...K) map[K]V  { return mapsimpl.Pick(m, keys...) }
func Omit[K comparable, V any](m map[K]V, keys ...K) map[K]V  { return mapsimpl.Omit(m, keys...) }
func Clear[K comparable, V any](m map[K]V)                    { mapsimpl.Clear(m) }
func Update[K comparable, V any](dst, src map[K]V) map[K]V    { return mapsimpl.Update(dst, src) }
func Clone[K comparable, V any](m map[K]V) map[K]V            { return mapsimpl.Clone(m) }
func Equal[K, V comparable](a, b map[K]V) bool                { return mapsimpl.Equal(a, b) }
func EqualFunc[K comparable, V1, V2 any](a map[K]V1, b map[K]V2, eq func(V1, V2) bool) bool {
	return mapsimpl.EqualFunc(a, b, eq)
}
