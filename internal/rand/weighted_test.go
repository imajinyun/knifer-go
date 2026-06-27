package rand

import (
	"errors"
	"math"
	mathrand "math/rand"
	"slices"
	"testing"

	knifer "github.com/imajinyun/knifer-go"
)

func TestWeightedPick(t *testing.T) {
	items := []string{"low", "high"}
	weights := []float64{0, 10}

	got, err := WeightedPick(
		items,
		weights,
		WithWeightedRandSource(mathrand.New(mathrand.NewSource(1))),
	)
	if err != nil {
		t.Fatalf("WeightedPick error = %v", err)
	}
	if got != "high" {
		t.Fatalf("WeightedPick = %q, want high", got)
	}
}

func TestWeightedPickN(t *testing.T) {
	items := []int{1, 2, 3}
	weights := []float64{1, 2, 3}

	got, err := WeightedPickN(
		items,
		weights,
		5,
		WithWeightedRandSource(mathrand.New(mathrand.NewSource(2))),
	)
	if err != nil {
		t.Fatalf("WeightedPickN error = %v", err)
	}
	if len(got) != 5 {
		t.Fatalf("WeightedPickN len = %d, want 5", len(got))
	}
	for _, item := range got {
		if !slices.Contains(items, item) {
			t.Fatalf("WeightedPickN item = %d, not in source items", item)
		}
	}
}

func TestWeightedPickUniqueN(t *testing.T) {
	got, err := WeightedPickUniqueN(
		[]string{"a", "b", "c"},
		[]float64{1, 1, 1},
		3,
		WithWeightedRandSource(mathrand.New(mathrand.NewSource(3))),
	)
	if err != nil {
		t.Fatalf("WeightedPickUniqueN error = %v", err)
	}
	if len(got) != 3 {
		t.Fatalf("WeightedPickUniqueN len = %d, want 3", len(got))
	}
	seen := map[string]bool{}
	for _, item := range got {
		if seen[item] {
			t.Fatalf("WeightedPickUniqueN repeated item %q in %v", item, got)
		}
		seen[item] = true
	}
}

func TestWeightedPickUniqueNZeroCount(t *testing.T) {
	got, err := WeightedPickUniqueN(
		[]string{"a"},
		[]float64{1},
		0,
		WithWeightedRandSource(mathrand.New(mathrand.NewSource(4))),
	)
	if err != nil {
		t.Fatalf("WeightedPickUniqueN zero count error = %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("WeightedPickUniqueN zero count = %v, want empty", got)
	}
}

func TestWeightedPickInvalidInput(t *testing.T) {
	tests := []struct {
		name string
		err  error
	}{
		{name: "empty items", err: weightedPickErr(WeightedPick([]string{}, []float64{}))},
		{name: "length mismatch", err: weightedPickErr(WeightedPick([]string{"a"}, []float64{}))},
		{name: "negative weight", err: weightedPickErr(WeightedPick([]string{"a"}, []float64{-1}))},
		{name: "nan weight", err: weightedPickErr(WeightedPick([]string{"a"}, []float64{math.NaN()}))},
		{name: "infinite weight", err: weightedPickErr(WeightedPick([]string{"a"}, []float64{math.Inf(1)}))},
		{name: "zero total", err: weightedPickErr(WeightedPick([]string{"a"}, []float64{0}))},
		{name: "negative count", err: weightedSliceErr(WeightedPickN([]string{"a"}, []float64{1}, -1))},
		{name: "negative unique count", err: weightedSliceErr(WeightedPickUniqueN([]string{"a"}, []float64{1}, -1))},
		{name: "unique count too large", err: weightedSliceErr(WeightedPickUniqueN([]string{"a"}, []float64{1}, 2))},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !errors.Is(tt.err, knifer.ErrCodeInvalidInput) {
				t.Fatalf("error = %v, want ErrCodeInvalidInput", tt.err)
			}
		})
	}
}

func TestWeightedOptionsAndDefensiveCopies(t *testing.T) {
	source := mathrand.New(mathrand.NewSource(5))
	cfg := applyWeightedOptions([]WeightedOption{
		nil,
		WithWeightedRandSource(source),
		WithWeightedPrecision(-1),
		WithWeightedPrecision(0.5),
	})
	if cfg.source != source {
		t.Fatalf("source option was not applied")
	}
	if cfg.precision != 0.5 {
		t.Fatalf("precision = %v, want 0.5", cfg.precision)
	}

	items := []string{"a", "b"}
	weights := []float64{1, 1}
	picker, err := newAliasPicker(items, weights, cfg)
	if err != nil {
		t.Fatalf("newAliasPicker error = %v", err)
	}
	items[0] = "mutated"
	if picker.items[0] != "a" {
		t.Fatalf("picker items were not defensively copied: %v", picker.items)
	}
}

func TestWeightedDefaultRandomSource(t *testing.T) {
	_, err := WeightedPick([]string{"a", "b"}, []float64{1, 1})
	if err != nil {
		t.Fatalf("WeightedPick with default source error = %v", err)
	}
}

func weightedPickErr[T any](_ T, err error) error {
	return err
}

func weightedSliceErr[T any](_ []T, err error) error {
	return err
}
