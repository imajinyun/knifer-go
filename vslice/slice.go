package vslice

import (
	"iter"

	sliceimpl "github.com/imajinyun/knifer-go/internal/slice"
)

func IsEmpty[T any](a []T) bool                { return sliceimpl.IsEmpty(a) }
func IsNotEmpty[T any](a []T) bool             { return sliceimpl.IsNotEmpty(a) }
func Contains[T comparable](a []T, v T) bool   { return sliceimpl.Contains(a, v) }
func IndexOf[T comparable](a []T, v T) int     { return sliceimpl.IndexOf(a, v) }
func LastIndexOf[T comparable](a []T, v T) int { return sliceimpl.LastIndexOf(a, v) }
func Reverse[T any](a []T) []T                 { return sliceimpl.Reverse(a) }
func Distinct[T comparable](a []T) []T         { return sliceimpl.Distinct(a) }
func Uniq[T comparable](a []T) []T             { return sliceimpl.Uniq(a) }
func UniqBy[T any, K comparable](a []T, keyFn func(T) K) []T {
	return sliceimpl.UniqBy(a, keyFn)
}
func Join[T any](a []T, sep string) string       { return sliceimpl.Join(a, sep) }
func Filter[T any](a []T, pred func(T) bool) []T { return sliceimpl.Filter(a, pred) }
func Reject[T any](a []T, pred func(T) bool) []T { return sliceimpl.Reject(a, pred) }
func FilterMap[T, R any](a []T, fn func(T) (R, bool)) []R {
	return sliceimpl.FilterMap(a, fn)
}
func Map[T, R any](a []T, fn func(T) R) []R { return sliceimpl.Map(a, fn) }
func MapErr[T, R any](a []T, fn func(T) (R, error)) ([]R, error) {
	return sliceimpl.MapErr(a, fn)
}
func FlatMap[T, R any](a []T, fn func(T) []R) []R { return sliceimpl.FlatMap(a, fn) }
func Reduce[T, R any](a []T, initial R, fn func(R, T) R) R {
	return sliceimpl.Reduce(a, initial, fn)
}

func FilterErr[T any](a []T, pred func(T) (bool, error)) ([]T, error) {
	return sliceimpl.FilterErr(a, pred)
}

func ReduceErr[T, R any](a []T, initial R, fn func(R, T) (R, error)) (R, error) {
	return sliceimpl.ReduceErr(a, initial, fn)
}
func ForEach[T any](a []T, fn func(T)) { sliceimpl.ForEach(a, fn) }
func Find[T any](a []T, pred func(T) bool) (T, bool) {
	return sliceimpl.Find(a, pred)
}
func FindIndex[T any](a []T, pred func(T) bool) int { return sliceimpl.FindIndex(a, pred) }
func Iter[T any](a []T) iter.Seq[T]                 { return sliceimpl.Iter(a) }
func IterIndexed[T any](a []T) iter.Seq2[int, T]    { return sliceimpl.IterIndexed(a) }
func GroupBy[T any, K comparable](a []T, keyFn func(T) K) map[K][]T {
	return sliceimpl.GroupBy(a, keyFn)
}

func CountBy[T any, K comparable](a []T, keyFn func(T) K) map[K]int {
	return sliceimpl.CountBy(a, keyFn)
}

func KeyBy[T any, K comparable](a []T, keyFn func(T) K) map[K]T {
	return sliceimpl.KeyBy(a, keyFn)
}

func Associate[T any, K comparable, V any](a []T, transform func(T) (K, V)) map[K]V {
	return sliceimpl.Associate(a, transform)
}

func SliceToMap[T any, K comparable, V any](a []T, transform func(T) (K, V)) map[K]V {
	return sliceimpl.SliceToMap(a, transform)
}
func Chunk[T any](a []T, size int) [][]T  { return sliceimpl.Chunk(a, size) }
func Window[T any](a []T, size int) [][]T { return sliceimpl.Window(a, size) }
func Sliding[T any](a []T, size, step int) [][]T {
	return sliceimpl.Sliding(a, size, step)
}

type Pair[A, B any] = sliceimpl.Pair[A, B]

func Zip2[A, B any](a []A, b []B) []Pair[A, B] { return sliceimpl.Zip2(a, b) }
func Unzip2[A, B any](pairs []Pair[A, B]) ([]A, []B) {
	return sliceimpl.Unzip2(pairs)
}
func Flatten[T any](a [][]T) []T      { return sliceimpl.Flatten(a) }
func Compact[T comparable](a []T) []T { return sliceimpl.Compact(a) }
func PartitionBy[T any, K comparable](a []T, keyFn func(T) K) [][]T {
	return sliceimpl.PartitionBy(a, keyFn)
}
func Sub[T any](a []T, fromIndex, toIndex int) []T { return sliceimpl.Sub(a, fromIndex, toIndex) }
func Concat[T any](slices ...[]T) []T              { return sliceimpl.Concat(slices...) }
func Union[T comparable](a, b []T) []T             { return sliceimpl.Union(a, b) }
func Intersection[T comparable](a, b []T) []T      { return sliceimpl.Intersection(a, b) }
func Subtract[T comparable](a, b []T) []T          { return sliceimpl.Subtract(a, b) }
func Page[T any](a []T, pageNo, pageSize int) []T  { return sliceimpl.Page(a, pageNo, pageSize) }
