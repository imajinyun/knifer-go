package job

import (
	"context"
	"errors"
	"io"
	"os"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	knifer "github.com/imajinyun/go-knifer"
	"github.com/sirupsen/logrus"
)

func TestMain(m *testing.M) {
	logrus.SetOutput(io.Discard)
	os.Exit(m.Run())
}

func TestRunUsesDefaultOptionsAsSingleSerialShard_BitsUT(t *testing.T) {
	var (
		ranges []string
		merged []string
	)

	j := NewSlice(func(ctx context.Context, start, end int) (Merge, error) {
		if ctx == nil {
			t.Fatal("ctx should not be nil")
		}
		ranges = append(ranges, formatRange(start, end))
		return func() error {
			merged = append(merged, formatRange(start, end))
			return nil
		}, nil
	}, 5)

	if err := Run(context.Background(), j); err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if want := []string{"0:5"}; !reflect.DeepEqual(ranges, want) {
		t.Fatalf("ranges = %v, want %v", ranges, want)
	}
	if want := []string{"0:5"}; !reflect.DeepEqual(merged, want) {
		t.Fatalf("merged = %v, want %v", merged, want)
	}
}

func TestRunWithUsesExplicitOptionsAndSerialMergeOrder_BitsUT(t *testing.T) {
	var (
		mu              sync.Mutex
		ranges          []string
		merged          []string
		active, maxSeen atomic.Int32
	)

	j := NewSlice(func(ctx context.Context, start, end int) (Merge, error) {
		current := active.Add(1)
		for {
			max := maxSeen.Load()
			if current <= max || maxSeen.CompareAndSwap(max, current) {
				break
			}
		}
		defer active.Add(-1)

		if start == 0 {
			time.Sleep(20 * time.Millisecond)
		}
		mu.Lock()
		ranges = append(ranges, formatRange(start, end))
		mu.Unlock()

		return func() error {
			merged = append(merged, formatRange(start, end))
			return nil
		}, nil
	}, 5)

	err := RunWith(context.Background(), j, Options{BatchSize: 2, MaxConcurrency: 2})
	if err != nil {
		t.Fatalf("RunWith() error = %v", err)
	}

	sort.Strings(ranges)
	if want := []string{"0:2", "2:4", "4:5"}; !reflect.DeepEqual(ranges, want) {
		t.Fatalf("ranges = %v, want %v", ranges, want)
	}
	if want := []string{"0:2", "2:4", "4:5"}; !reflect.DeepEqual(merged, want) {
		t.Fatalf("merged = %v, want %v", merged, want)
	}
	if got := maxSeen.Load(); got > 2 {
		t.Fatalf("max concurrency = %d, want <= 2", got)
	}
}

func TestRunWithEmbeddedOptions_BitsUT(t *testing.T) {
	wrapped := &embeddedOptionsJob{
		Options: Options{BatchSize: 2, MaxConcurrency: 2},
		vals:    []int{1, 2, 3, 4},
	}
	if err := RunWith(context.Background(), wrapped, wrapped.Options); err != nil {
		t.Fatalf("RunWith() error = %v", err)
	}
	sort.Ints(wrapped.seen)
	if want := []int{1, 2, 3, 4}; !reflect.DeepEqual(wrapped.seen, want) {
		t.Fatalf("seen = %v, want %v", wrapped.seen, want)
	}
}

type embeddedOptionsJob struct {
	Options
	vals []int
	seen []int
}

func (j *embeddedOptionsJob) Len() int { return len(j.vals) }

func (j *embeddedOptionsJob) Run(ctx context.Context, start, end int) (Merge, error) {
	_ = ctx
	batch := append([]int(nil), j.vals[start:end]...)
	return func() error {
		j.seen = append(j.seen, batch...)
		return nil
	}, nil
}

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

func TestBatchSliceAndMapAdapters_BitsUT(t *testing.T) {
	t.Run("batch", func(t *testing.T) {
		var seen []int
		j := NewBatch(func(ctx context.Context, vals []int) (Merge, error) {
			copied := append([]int(nil), vals...)
			return func() error {
				seen = append(seen, copied...)
				return nil
			}, nil
		}, []int{1, 2, 3}).WithBatchSize(2)

		if err := RunWith(context.Background(), j, j.Options); err != nil {
			t.Fatalf("RunWith() error = %v", err)
		}
		sort.Ints(seen)
		if want := []int{1, 2, 3}; !reflect.DeepEqual(seen, want) {
			t.Fatalf("seen = %v, want %v", seen, want)
		}
	})

	t.Run("single", func(t *testing.T) {
		var sum int
		j := NewBatchSingle(func(ctx context.Context, v int) (Merge, error) {
			return func() error {
				sum += v
				return nil
			}, nil
		}, []int{1, 2, 3})

		if err := RunWith(context.Background(), j, j.Options); err != nil {
			t.Fatalf("RunWith() error = %v", err)
		}
		if sum != 6 {
			t.Fatalf("sum = %d, want 6", sum)
		}
	})

	t.Run("map keys", func(t *testing.T) {
		data := map[string]int{"b": 2, "a": 1}
		var keys []string
		j := NewMapKeys(func(ctx context.Context, key string) (Merge, error) {
			return func() error {
				keys = append(keys, key)
				return nil
			}, nil
		}, data)

		if err := RunWith(context.Background(), j, j.Options); err != nil {
			t.Fatalf("RunWith() error = %v", err)
		}
		sort.Strings(keys)
		if want := []string{"a", "b"}; !reflect.DeepEqual(keys, want) {
			t.Fatalf("keys = %v, want %v", keys, want)
		}
	})

	t.Run("reflect map", func(t *testing.T) {
		data := map[int]string{2: "b", 1: "a"}
		var keys []int
		j := NewMap(func(ctx context.Context, key int) (Merge, error) {
			return func() error {
				keys = append(keys, key)
				return nil
			}, nil
		}, data)

		if err := RunWith(context.Background(), j, j.Options); err != nil {
			t.Fatalf("RunWith() error = %v", err)
		}
		sort.Ints(keys)
		if want := []int{1, 2}; !reflect.DeepEqual(keys, want) {
			t.Fatalf("keys = %v, want %v", keys, want)
		}
	})
}

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
		expectPanic(t, func() { NewMap(123, map[string]int{"a": 1}) })
		expectPanic(t, func() { NewMap(func(context.Context, string) error { return nil }, map[string]int{"a": 1}) })
		expectPanic(t, func() {
			NewMap(func(context.Context, string) (error, error) { return nil, nil }, map[string]int{"a": 1})
		})
		expectPanic(t, func() { NewMap(func(context.Context, string) (Merge, error) { return nil, nil }, []string{"a"}) })
		expectPanic(t, func() { NewMap(func(context.Context, int) (Merge, error) { return nil, nil }, map[string]int{"a": 1}) })
	})
}

func TestOptionsAndHelpers_BitsUT(t *testing.T) {
	if got := (Options{}).normalized(7); got.BatchSize != 7 || got.MaxConcurrency != 1 {
		t.Fatalf("Options{}.normalized(7) = %+v, want batch 7 concurrency 1", got)
	}
	if got := (Options{BatchSize: 3, MaxConcurrency: 2}).normalized(7); got.BatchSize != 3 || got.MaxConcurrency != 2 {
		t.Fatalf("Options.normalized(7) = %+v, want batch 3 concurrency 2", got)
	}
	if got := chunks(7, 3); got != 3 {
		t.Fatalf("chunks(7, 3) = %d, want 3", got)
	}
	if got := chunks(7, 0); got != 0 {
		t.Fatalf("chunks(7, 0) = %d, want 0", got)
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

func formatRange(start, end int) string {
	return strings.Join([]string{itoa(start), itoa(end)}, ":")
}

func itoa(v int) string {
	return strconv.Itoa(v)
}
