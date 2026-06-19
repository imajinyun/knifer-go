// Package slice provides slice helpers.
package slice

import (
	"fmt"
	"iter"
	"slices"
	"strings"
)

// This file provides generic slice helpers aligned with the utility toolkit-core ArrayUtil.
// Functions return new slices where mutation would be surprising, while
// Reverse intentionally reverses the input slice in place for efficiency.

// IsEmpty reports whether the slice is empty.
func IsEmpty[T any](a []T) bool { return len(a) == 0 }

// IsNotEmpty reports whether the slice is not empty.
func IsNotEmpty[T any](a []T) bool { return len(a) > 0 }

// Contains reports whether the slice contains v. T must be comparable.
func Contains[T comparable](a []T, v T) bool {
	return slices.Contains(a, v)
}

// IndexOf returns the first index of v, or -1 when v is absent.
func IndexOf[T comparable](a []T, v T) int {
	return slices.Index(a, v)
}

// LastIndexOf returns the last index of v, or -1 when v is absent.
func LastIndexOf[T comparable](a []T, v T) int {
	for i := len(a) - 1; i >= 0; i-- {
		if a[i] == v {
			return i
		}
	}
	return -1
}

// Reverse reverses the input slice in place and returns the same slice.
func Reverse[T any](a []T) []T {
	slices.Reverse(a)
	return a
}

// Distinct removes duplicates while preserving the first occurrence order.
func Distinct[T comparable](a []T) []T {
	return Uniq(a)
}

// Uniq removes duplicates while preserving the first occurrence order.
func Uniq[T comparable](a []T) []T {
	seen := make(map[T]struct{}, len(a))
	out := make([]T, 0, len(a))
	for _, v := range a {
		if _, ok := seen[v]; ok {
			continue
		}
		seen[v] = struct{}{}
		out = append(out, v)
	}
	return out
}

// UniqBy removes duplicates by key while preserving the first occurrence order.
func UniqBy[T any, K comparable](a []T, keyFn func(T) K) []T {
	seen := make(map[K]struct{}, len(a))
	out := make([]T, 0, len(a))
	for _, v := range a {
		key := keyFn(v)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, v)
	}
	return out
}

// Join converts elements with fmt.Sprint and joins them with sep.
func Join[T any](a []T, sep string) string {
	parts := make([]string, len(a))
	for i, v := range a {
		parts[i] = fmt.Sprint(v)
	}
	return strings.Join(parts, sep)
}

// Filter returns elements for which pred returns true.
func Filter[T any](a []T, pred func(T) bool) []T {
	out := make([]T, 0, len(a))
	for _, v := range a {
		if pred(v) {
			out = append(out, v)
		}
	}
	return out
}

// Reject returns elements for which pred returns false.
func Reject[T any](a []T, pred func(T) bool) []T {
	return Filter(a, func(v T) bool { return !pred(v) })
}

// FilterMap transforms elements and keeps only values explicitly accepted by fn.
func FilterMap[T, R any](a []T, fn func(T) (R, bool)) []R {
	out := make([]R, 0, len(a))
	for _, v := range a {
		mapped, ok := fn(v)
		if ok {
			out = append(out, mapped)
		}
	}
	return out
}

// Map maps each element to another value while preserving order.
func Map[T, R any](a []T, fn func(T) R) []R {
	out := make([]R, len(a))
	for i, v := range a {
		out[i] = fn(v)
	}
	return out
}

// MapErr maps each element while preserving order and stops on the first error.
// The returned slice contains values produced before the failing callback.
func MapErr[T, R any](a []T, fn func(T) (R, error)) ([]R, error) {
	out := make([]R, 0, len(a))
	for _, v := range a {
		mapped, err := fn(v)
		if err != nil {
			return out, err
		}
		out = append(out, mapped)
	}
	return out, nil
}

// FlatMap maps each element to zero or more values and flattens one level.
func FlatMap[T, R any](a []T, fn func(T) []R) []R {
	out := make([]R, 0, len(a))
	for _, v := range a {
		out = append(out, fn(v)...)
	}
	return out
}

// Reduce folds a slice from left to right.
func Reduce[T, R any](a []T, initial R, fn func(R, T) R) R {
	acc := initial
	for _, v := range a {
		acc = fn(acc, v)
	}
	return acc
}

// FilterErr returns elements for which pred returns true and stops on the first error.
// The returned slice contains accepted values produced before the failing callback.
func FilterErr[T any](a []T, pred func(T) (bool, error)) ([]T, error) {
	out := make([]T, 0, len(a))
	for _, v := range a {
		keep, err := pred(v)
		if err != nil {
			return out, err
		}
		if keep {
			out = append(out, v)
		}
	}
	return out, nil
}

// ReduceErr folds a slice from left to right and stops on the first error.
// The returned accumulator is the last successful accumulator value.
func ReduceErr[T, R any](a []T, initial R, fn func(R, T) (R, error)) (R, error) {
	acc := initial
	for _, v := range a {
		next, err := fn(acc, v)
		if err != nil {
			return acc, err
		}
		acc = next
	}
	return acc, nil
}

// ForEach invokes fn for every element in order.
func ForEach[T any](a []T, fn func(T)) {
	for _, v := range a {
		fn(v)
	}
}

// Find returns the first element satisfying pred.
func Find[T any](a []T, pred func(T) bool) (T, bool) {
	idx := slices.IndexFunc(a, pred)
	if idx >= 0 {
		return a[idx], true
	}
	var zero T
	return zero, false
}

// FindIndex returns the first index satisfying pred, or -1 when absent.
func FindIndex[T any](a []T, pred func(T) bool) int {
	return slices.IndexFunc(a, pred)
}

// Iter returns an iterator over slice values in index order.
func Iter[T any](a []T) iter.Seq[T] { return slices.Values(a) }

// IterIndexed returns an iterator over slice index-value pairs in index order.
func IterIndexed[T any](a []T) iter.Seq2[int, T] { return slices.All(a) }

// GroupBy groups slice items by keyFn while preserving item order inside each group.
func GroupBy[T any, K comparable](a []T, keyFn func(T) K) map[K][]T {
	out := make(map[K][]T)
	for _, v := range a {
		key := keyFn(v)
		out[key] = append(out[key], v)
	}
	return out
}

// CountBy counts slice items grouped by keyFn.
func CountBy[T any, K comparable](a []T, keyFn func(T) K) map[K]int {
	out := make(map[K]int)
	for _, v := range a {
		out[keyFn(v)]++
	}
	return out
}

// KeyBy builds a map from keyFn(item) to item. Later duplicate keys overwrite earlier ones.
func KeyBy[T any, K comparable](a []T, keyFn func(T) K) map[K]T {
	out := make(map[K]T, len(a))
	for _, v := range a {
		out[keyFn(v)] = v
	}
	return out
}

// Associate builds a map from transform(item). Later duplicate keys overwrite earlier ones.
func Associate[T any, K comparable, V any](a []T, transform func(T) (K, V)) map[K]V {
	out := make(map[K]V, len(a))
	for _, item := range a {
		key, value := transform(item)
		out[key] = value
	}
	return out
}

// SliceToMap is an alias of Associate for callers familiar with lo-style naming.
func SliceToMap[T any, K comparable, V any](a []T, transform func(T) (K, V)) map[K]V {
	return Associate(a, transform)
}

// Chunk splits a slice into fixed-size chunks. Non-positive size returns an empty slice.
func Chunk[T any](a []T, size int) [][]T {
	if size <= 0 || len(a) == 0 {
		return [][]T{}
	}
	out := make([][]T, 0, (len(a)+size-1)/size)
	for start := 0; start < len(a); start += size {
		end := start + size
		if end > len(a) {
			end = len(a)
		}
		out = append(out, slices.Clone(a[start:end]))
	}
	return out
}

// Window returns overlapping fixed-size windows with a step of one.
// Non-positive size or size greater than len(a) returns an empty slice.
func Window[T any](a []T, size int) [][]T {
	return Sliding(a, size, 1)
}

// Sliding returns fixed-size windows advanced by step elements.
// Non-positive size or step, or size greater than len(a), returns an empty slice.
func Sliding[T any](a []T, size, step int) [][]T {
	if size <= 0 || step <= 0 || size > len(a) {
		return [][]T{}
	}
	out := make([][]T, 0, (len(a)-size)/step+1)
	for start := 0; start+size <= len(a); start += step {
		out = append(out, slices.Clone(a[start:start+size]))
	}
	return out
}

// Pair stores two typed values for Zip2 and Unzip2.
type Pair[A, B any] struct {
	First  A
	Second B
}

// Zip2 pairs elements from two slices up to the shorter length.
func Zip2[A, B any](a []A, b []B) []Pair[A, B] {
	n := min(len(a), len(b))
	out := make([]Pair[A, B], 0, n)
	for i := 0; i < n; i++ {
		out = append(out, Pair[A, B]{First: a[i], Second: b[i]})
	}
	return out
}

// Unzip2 splits pairs into two slices while preserving pair order.
func Unzip2[A, B any](pairs []Pair[A, B]) ([]A, []B) {
	left := make([]A, 0, len(pairs))
	right := make([]B, 0, len(pairs))
	for _, pair := range pairs {
		left = append(left, pair.First)
		right = append(right, pair.Second)
	}
	return left, right
}

// Flatten flattens one level of nested slices.
func Flatten[T any](a [][]T) []T {
	total := 0
	for _, item := range a {
		total += len(item)
	}
	out := make([]T, 0, total)
	for _, item := range a {
		out = append(out, item...)
	}
	return out
}

// Compact removes zero-value elements while preserving order.
func Compact[T comparable](a []T) []T {
	var zero T
	out := slices.Clone(a)
	out = slices.DeleteFunc(out, func(v T) bool { return v == zero })
	if out == nil {
		return []T{}
	}
	return out
}

// PartitionBy groups adjacent items that share the same key returned by keyFn.
func PartitionBy[T any, K comparable](a []T, keyFn func(T) K) [][]T {
	if len(a) == 0 {
		return [][]T{}
	}
	out := make([][]T, 0)
	current := []T{a[0]}
	currentKey := keyFn(a[0])
	for _, v := range a[1:] {
		key := keyFn(v)
		if key == currentKey {
			current = append(current, v)
			continue
		}
		out = append(out, current)
		current = []T{v}
		currentKey = key
	}
	return append(out, current)
}

// Sub returns a copied sub-slice and supports negative indexes.
// Negative indexes are resolved from the end of the slice, and reversed ranges
// are normalized by swapping fromIndex and toIndex, following the utility toolkit behavior.
func Sub[T any](a []T, fromIndex, toIndex int) []T {
	n := len(a)
	if n == 0 {
		return []T{}
	}
	if fromIndex < 0 {
		fromIndex += n
	}
	if toIndex < 0 {
		toIndex += n
	}
	if fromIndex < 0 {
		fromIndex = 0
	}
	if toIndex > n {
		toIndex = n
	}
	if fromIndex > toIndex {
		fromIndex, toIndex = toIndex, fromIndex
	}
	if fromIndex >= n {
		return []T{}
	}
	return slices.Clone(a[fromIndex:toIndex])
}

// Concat concatenates multiple slices into a new slice.
func Concat[T any](items ...[]T) []T {
	out := slices.Concat(items...)
	if out == nil {
		return []T{}
	}
	return out
}
