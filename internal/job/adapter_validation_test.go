package job

import (
	"context"
	"errors"
	"testing"

	knifer "github.com/imajinyun/knifer-go"
)

func TestAdaptersValidateRangesAndInputs_BitsUT(t *testing.T) {
	t.Run("invalid slice range", func(t *testing.T) {
		_, err := NewSlice(func(ctx context.Context, start, end int) (Merge, error) { return nil, nil }, 1).Run(context.Background(), 1, 2)
		if !errors.Is(err, ErrInvalidRange) {
			t.Fatalf("Slice.Run() error = %v, want ErrInvalidRange", err)
		}
		if !errors.Is(err, knifer.ErrCodeInvalidInput) {
			t.Fatalf("Slice.Run() error = %v, want ErrCodeInvalidInput", err)
		}
	})

	t.Run("invalid batch range", func(t *testing.T) {
		_, err := NewBatch(func(ctx context.Context, vals []int) (Merge, error) { return nil, nil }, []int{1}).Run(context.Background(), -1, 1)
		if !errors.Is(err, ErrInvalidRange) {
			t.Fatalf("Batch.Run() error = %v, want ErrInvalidRange", err)
		}
	})

	t.Run("invalid reflect map input", func(t *testing.T) {
		tests := []struct {
			name string
			run  any
			data any
		}{
			{name: "run is not func", run: 123, data: map[string]int{"a": 1}},
			{name: "invalid signature", run: func(context.Context, string) error { return nil }, data: map[string]int{"a": 1}},
			{name: "invalid return", run: func(context.Context, string) (error, error) { return nil, nil }, data: map[string]int{"a": 1}},
			{name: "data is not map", run: func(context.Context, string) (Merge, error) { return nil, nil }, data: []string{"a"}},
			{name: "key type mismatch", run: func(context.Context, int) (Merge, error) { return nil, nil }, data: map[string]int{"a": 1}},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				if _, err := NewMapE(tt.run, tt.data); !errors.Is(err, ErrInvalidMapJob) {
					t.Fatalf("NewMapE() error = %v, want ErrInvalidMapJob", err)
				}
				expectPanic(t, func() { NewMap(tt.run, tt.data) })
			})
		}
	})
}
