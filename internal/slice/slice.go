// Package slice provides slice helpers.
package slice

import (
	"fmt"
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
	for _, x := range a {
		if x == v {
			return true
		}
	}
	return false
}

// IndexOf returns the first index of v, or -1 when v is absent.
func IndexOf[T comparable](a []T, v T) int {
	for i, x := range a {
		if x == v {
			return i
		}
	}
	return -1
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
	for i, j := 0, len(a)-1; i < j; i, j = i+1, j-1 {
		a[i], a[j] = a[j], a[i]
	}
	return a
}

// Distinct removes duplicates while preserving the first occurrence order.
func Distinct[T comparable](a []T) []T {
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

// Map maps each element to another value while preserving order.
func Map[T, R any](a []T, fn func(T) R) []R {
	out := make([]R, len(a))
	for i, v := range a {
		out[i] = fn(v)
	}
	return out
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
	out := make([]T, toIndex-fromIndex)
	copy(out, a[fromIndex:toIndex])
	return out
}

// Concat concatenates multiple slices into a new slice.
func Concat[T any](slices ...[]T) []T {
	total := 0
	for _, s := range slices {
		total += len(s)
	}
	out := make([]T, 0, total)
	for _, s := range slices {
		out = append(out, s...)
	}
	return out
}
