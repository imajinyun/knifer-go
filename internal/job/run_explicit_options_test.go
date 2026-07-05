package job

import (
	"context"
	"errors"
	"reflect"
	"slices"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

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

	slices.Sort(ranges)
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

func TestRunWithConcurrentErrorCancelsSiblingShard(t *testing.T) {
	wantErr := errors.New("stop shard")
	startedSlow := make(chan struct{})
	releaseFast := make(chan struct{})
	slowCanceled := make(chan struct{}, 1)

	j := NewSlice(func(ctx context.Context, start, end int) (Merge, error) {
		switch start {
		case 0:
			close(startedSlow)
			<-ctx.Done()
			slowCanceled <- struct{}{}
			return nil, ctx.Err()
		case 1:
			<-startedSlow
			<-releaseFast
			return nil, wantErr
		default:
			return nil, nil
		}
	}, 2)

	done := make(chan error, 1)
	go func() {
		done <- RunWith(context.Background(), j, Options{BatchSize: 1, MaxConcurrency: 2})
	}()

	select {
	case <-startedSlow:
	case <-time.After(time.Second):
		t.Fatal("slow shard did not start")
	}
	close(releaseFast)

	select {
	case err := <-done:
		if err == nil || !strings.Contains(err.Error(), wantErr.Error()) {
			t.Fatalf("RunWith() error = %v, want %q", err, wantErr.Error())
		}
	case <-time.After(time.Second):
		t.Fatal("RunWith did not return after shard error")
	}
	select {
	case <-slowCanceled:
	default:
		t.Fatal("sibling shard did not observe canceled context")
	}
}
