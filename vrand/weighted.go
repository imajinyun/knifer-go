package vrand

import (
	mathrand "math/rand"

	randimpl "github.com/imajinyun/knifer-go/internal/rand"
)

// WeightedOption customizes weighted random selection.
type WeightedOption = randimpl.WeightedOption

// WithWeightedRandSource sets the pseudo-random source used by weighted helpers.
func WithWeightedRandSource(source *mathrand.Rand) WeightedOption {
	return randimpl.WithWeightedRandSource(source)
}

// WithWeightedPrecision sets the minimum positive total-weight tolerance.
func WithWeightedPrecision(precision float64) WeightedOption {
	return randimpl.WithWeightedPrecision(precision)
}

// WeightedPick picks one item according to weights.
func WeightedPick[T any](items []T, weights []float64, opts ...WeightedOption) (T, error) {
	return randimpl.WeightedPick(items, weights, opts...)
}

// WeightedPickN picks n items with replacement according to weights.
func WeightedPickN[T any](items []T, weights []float64, n int, opts ...WeightedOption) ([]T, error) {
	return randimpl.WeightedPickN(items, weights, n, opts...)
}

// WeightedPickUniqueN picks n unique items without replacement according to weights.
func WeightedPickUniqueN[T any](items []T, weights []float64, n int, opts ...WeightedOption) ([]T, error) {
	return randimpl.WeightedPickUniqueN(items, weights, n, opts...)
}
