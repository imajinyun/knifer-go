package vjob

import (
	"context"
	"reflect"
	"slices"
	"testing"
)

func TestFacadeNewSliceSingle(t *testing.T) {
	var indices []int
	j := NewSliceSingle(func(ctx context.Context, idx int) (Merge, error) {
		return func() error {
			indices = append(indices, idx)
			return nil
		}, nil
	}, 3)
	if err := Run(context.Background(), j); err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	slices.Sort(indices)
	if want := []int{0, 1, 2}; !reflect.DeepEqual(indices, want) {
		t.Fatalf("indices = %v, want %v", indices, want)
	}
}

func TestFacadeNewBatchSingle(t *testing.T) {
	var seen []int
	j := NewBatchSingle(func(ctx context.Context, val int) (Merge, error) {
		return func() error {
			seen = append(seen, val)
			return nil
		}, nil
	}, []int{3, 1, 2}).WithBatchSize(1).WithMaxConcurrency(2)
	if err := Run(context.Background(), j); err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	slices.Sort(seen)
	if want := []int{1, 2, 3}; !reflect.DeepEqual(seen, want) {
		t.Fatalf("seen = %v, want %v", seen, want)
	}
}

func TestFacadeNewMap(t *testing.T) {
	data := map[int]string{2: "b", 1: "a"}
	var keys []int
	j := NewMap(func(ctx context.Context, key int) (Merge, error) {
		return func() error {
			keys = append(keys, key)
			return nil
		}, nil
	}, data)
	if err := Run(context.Background(), j); err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	slices.Sort(keys)
	if want := []int{1, 2}; !reflect.DeepEqual(keys, want) {
		t.Fatalf("keys = %v, want %v", keys, want)
	}
}
