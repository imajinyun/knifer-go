package vrand

import (
	"errors"
	mathrand "math/rand"
	"testing"

	knifer "github.com/imajinyun/knifer-go"
)

func TestWeightedPickFacade(t *testing.T) {
	items := []string{"a", "b", "c"}
	weights := []float64{0, 1, 0}
	got, err := WeightedPick(items, weights, WithWeightedRandSource(mathrand.New(mathrand.NewSource(1))))
	if err != nil || got != "b" {
		t.Fatalf("WeightedPick = %q, %v", got, err)
	}
	many, err := WeightedPickN(items, weights, 3, WithWeightedRandSource(mathrand.New(mathrand.NewSource(1))))
	if err != nil || len(many) != 3 {
		t.Fatalf("WeightedPickN = %v, %v", many, err)
	}
	for _, item := range many {
		if item != "b" {
			t.Fatalf("WeightedPickN item = %q", item)
		}
	}
}

func TestWeightedPickUniqueNFacade(t *testing.T) {
	got, err := WeightedPickUniqueN(
		[]string{"a", "b", "c"},
		[]float64{1, 1, 1},
		2,
		WithWeightedRandSource(mathrand.New(mathrand.NewSource(2))),
	)
	if err != nil {
		t.Fatalf("WeightedPickUniqueN error = %v", err)
	}
	if len(got) != 2 || got[0] == got[1] {
		t.Fatalf("WeightedPickUniqueN = %v", got)
	}
}

func TestWeightedPickInvalidInput(t *testing.T) {
	tests := []struct {
		name string
		err  error
	}{
		{name: "empty", err: weightedErr(WeightedPick([]string{}, []float64{}))},
		{name: "mismatch", err: weightedErr(WeightedPick([]string{"a"}, []float64{}))},
		{name: "negative", err: weightedErr(WeightedPick([]string{"a"}, []float64{-1}))},
		{name: "zero", err: weightedErr(WeightedPick([]string{"a"}, []float64{0}))},
		{name: "unique count", err: weightedSliceErr(WeightedPickUniqueN([]string{"a"}, []float64{1}, 2))},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !errors.Is(tt.err, knifer.ErrCodeInvalidInput) {
				t.Fatalf("err = %v, want invalid input", tt.err)
			}
		})
	}
}

func weightedErr[T any](_ T, err error) error        { return err }
func weightedSliceErr[T any](_ []T, err error) error { return err }
