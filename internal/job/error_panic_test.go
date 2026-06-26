package job

import (
	"context"
	"errors"
	"strings"
	"testing"

	knifer "github.com/imajinyun/knifer-go"
)

func TestRunWithHandlesErrorsAndPanics_BitsUT(t *testing.T) {
	t.Run("nil job", func(t *testing.T) {
		if err := RunWith(context.Background(), nil, Options{}); !errors.Is(err, ErrNilJob) {
			t.Fatalf("RunWith(nil) error = %v, want ErrNilJob", err)
		}
		if err := RunWith(context.Background(), nil, Options{}); !errors.Is(err, knifer.ErrCodeInvalidInput) {
			t.Fatalf("RunWith(nil) error = %v, want ErrCodeInvalidInput", err)
		}
	})

	t.Run("canceled context", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		j := NewSlice(func(ctx context.Context, start, end int) (Merge, error) {
			t.Fatal("Run should not be called when context is canceled")
			return nil, nil
		}, 1)
		if err := RunWith(ctx, j, Options{}); !errors.Is(err, context.Canceled) {
			t.Fatalf("RunWith(canceled) error = %v, want context.Canceled", err)
		}
	})

	t.Run("run error", func(t *testing.T) {
		wantErr := errors.New("run failed")
		j := NewSlice(func(ctx context.Context, start, end int) (Merge, error) {
			if start == 2 {
				return nil, wantErr
			}
			return nil, nil
		}, 4)
		if err := RunWith(context.Background(), j, Options{BatchSize: 2, MaxConcurrency: 1}); err == nil || !strings.Contains(err.Error(), wantErr.Error()) {
			t.Fatalf("RunWith() error = %v, want to contain %q", err, wantErr.Error())
		}
	})

	t.Run("merge panic", func(t *testing.T) {
		j := NewSlice(func(ctx context.Context, start, end int) (Merge, error) {
			return func() error { panic("merge boom") }, nil
		}, 1)
		if err := RunWith(context.Background(), j, Options{}); err == nil || !strings.Contains(err.Error(), "merge boom") {
			t.Fatalf("RunWith() error = %v, want recovered panic", err)
		}
	})
}

func TestSentinelErrorCode(t *testing.T) {
	if got := ErrNilJob.(*sentinel).ErrorCode(); got != knifer.ErrCodeInvalidInput {
		t.Fatalf("ErrNilJob.ErrorCode() = %v, want %v", got, knifer.ErrCodeInvalidInput)
	}
	if got := ErrInvalidRange.(*sentinel).ErrorCode(); got != knifer.ErrCodeInvalidInput {
		t.Fatalf("ErrInvalidRange.ErrorCode() = %v, want %v", got, knifer.ErrCodeInvalidInput)
	}
}

func TestBatchJobOptions(t *testing.T) {
	j := NewBatch(func(ctx context.Context, vals []int) (Merge, error) { return nil, nil }, []int{1, 2, 3})
	opts := j.JobOptions()
	if opts.BatchSize != 0 {
		t.Fatalf("JobOptions().BatchSize = %d, want 0", opts.BatchSize)
	}
	if opts.MaxConcurrency != 0 {
		t.Fatalf("JobOptions().MaxConcurrency = %d, want 0", opts.MaxConcurrency)
	}
}

func expectPanic(t *testing.T, fn func()) {
	t.Helper()
	defer func() {
		if recover() == nil {
			t.Fatal("expected panic")
		}
	}()
	fn()
}
