package rand

import (
	"math"
	mathrand "math/rand"

	knifer "github.com/imajinyun/knifer-go"
)

type weightedConfig struct {
	source    *mathrand.Rand
	precision float64
}

// WeightedOption customizes weighted random selection.
type WeightedOption func(*weightedConfig)

// WithWeightedRandSource sets the pseudo-random source used by weighted helpers.
func WithWeightedRandSource(source *mathrand.Rand) WeightedOption {
	return func(c *weightedConfig) {
		if source != nil {
			c.source = source
		}
	}
}

// WithWeightedPrecision sets the minimum positive total-weight tolerance.
func WithWeightedPrecision(precision float64) WeightedOption {
	return func(c *weightedConfig) {
		if precision > 0 {
			c.precision = precision
		}
	}
}

// WeightedPick picks one item according to weights.
func WeightedPick[T any](items []T, weights []float64, opts ...WeightedOption) (T, error) {
	picker, err := newAliasPicker(items, weights, applyWeightedOptions(opts))
	if err != nil {
		var zero T
		return zero, err
	}
	return picker.pick(), nil
}

// WeightedPickN picks n items with replacement according to weights.
func WeightedPickN[T any](items []T, weights []float64, n int, opts ...WeightedOption) ([]T, error) {
	if n < 0 {
		return nil, weightedInvalidInput("weighted pick count must be non-negative")
	}
	picker, err := newAliasPicker(items, weights, applyWeightedOptions(opts))
	if err != nil {
		return nil, err
	}
	out := make([]T, 0, n)
	for i := 0; i < n; i++ {
		out = append(out, picker.pick())
	}
	return out, nil
}

// WeightedPickUniqueN picks n unique items without replacement according to weights.
func WeightedPickUniqueN[T any](items []T, weights []float64, n int, opts ...WeightedOption) ([]T, error) {
	if n < 0 {
		return nil, weightedInvalidInput("weighted pick count must be non-negative")
	}
	if n > len(items) {
		return nil, weightedInvalidInput("weighted unique pick count exceeds item count")
	}
	cfg := applyWeightedOptions(opts)
	remainingItems := append([]T(nil), items...)
	remainingWeights := append([]float64(nil), weights...)
	out := make([]T, 0, n)
	for i := 0; i < n; i++ {
		picker, err := newAliasPicker(remainingItems, remainingWeights, cfg)
		if err != nil {
			return nil, err
		}
		idx := picker.pickIndex()
		out = append(out, remainingItems[idx])
		remainingItems = append(remainingItems[:idx], remainingItems[idx+1:]...)
		remainingWeights = append(remainingWeights[:idx], remainingWeights[idx+1:]...)
	}
	return out, nil
}

type aliasPicker[T any] struct {
	items []T
	prob  []float64
	alias []int
	cfg   weightedConfig
}

func newAliasPicker[T any](items []T, weights []float64, cfg weightedConfig) (*aliasPicker[T], error) {
	if len(items) == 0 {
		return nil, weightedInvalidInput("weighted pick items are empty")
	}
	if len(items) != len(weights) {
		return nil, weightedInvalidInput("weighted pick items and weights length mismatch")
	}

	total := 0.0
	for _, weight := range weights {
		if math.IsNaN(weight) || math.IsInf(weight, 0) || weight < 0 {
			return nil, weightedInvalidInput("weighted pick weights must be finite and non-negative")
		}
		total += weight
	}
	if total <= cfg.precision {
		return nil, weightedInvalidInput("weighted pick total weight must be positive")
	}

	n := len(items)
	scaled := make([]float64, n)
	small := make([]int, 0, n)
	large := make([]int, 0, n)
	for i, weight := range weights {
		scaled[i] = weight * float64(n) / total
		if scaled[i] < 1 {
			small = append(small, i)
		} else {
			large = append(large, i)
		}
	}

	prob := make([]float64, n)
	alias := make([]int, n)
	for len(small) > 0 && len(large) > 0 {
		s := small[len(small)-1]
		small = small[:len(small)-1]
		l := large[len(large)-1]
		large = large[:len(large)-1]

		prob[s] = scaled[s]
		alias[s] = l
		scaled[l] = scaled[l] + scaled[s] - 1
		if scaled[l] < 1 {
			small = append(small, l)
		} else {
			large = append(large, l)
		}
	}
	for _, i := range append(small, large...) {
		prob[i] = 1
		alias[i] = i
	}

	return &aliasPicker[T]{
		items: append([]T(nil), items...),
		prob:  prob,
		alias: alias,
		cfg:   cfg,
	}, nil
}

func (p *aliasPicker[T]) pick() T {
	return p.items[p.pickIndex()]
}

func (p *aliasPicker[T]) pickIndex() int {
	idx := weightedIntn(p.cfg, len(p.items))
	if weightedFloat64(p.cfg) < p.prob[idx] {
		return idx
	}
	return p.alias[idx]
}

func applyWeightedOptions(opts []WeightedOption) weightedConfig {
	cfg := weightedConfig{precision: 1e-12}
	for _, opt := range opts {
		if opt != nil {
			opt(&cfg)
		}
	}
	return cfg
}

func weightedIntn(cfg weightedConfig, n int) int {
	if cfg.source != nil {
		return cfg.source.Intn(n)
	}
	defaultRandMu.Lock()
	defer defaultRandMu.Unlock()
	return defaultRandLocked().Intn(n)
}

func weightedFloat64(cfg weightedConfig) float64 {
	if cfg.source != nil {
		return cfg.source.Float64()
	}
	defaultRandMu.Lock()
	defer defaultRandMu.Unlock()
	return defaultRandLocked().Float64()
}

func weightedInvalidInput(msg string) error {
	return &knifer.Error{Code: knifer.ErrCodeInvalidInput, Message: msg}
}
