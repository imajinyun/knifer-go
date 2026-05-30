package vsem_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/imajinyun/go-knifer/vsem"
)

func TestVSemFacade(t *testing.T) {
	sem := vsem.New(1)
	if sem.Cap() != 1 {
		t.Fatalf("Cap() = %d, want 1", sem.Cap())
	}
	if err := sem.Acquire(context.Background(), 1); err != nil {
		t.Fatal(err)
	}
	if sem.InUse() != 1 {
		t.Fatalf("InUse() = %d, want 1", sem.InUse())
	}
	if sem.TryAcquire(1) {
		t.Fatal("TryAcquire() should fail when capacity is full")
	}
	sem.Release(1)
	if !sem.TryAcquire(1) {
		t.Fatal("TryAcquire() should succeed after release")
	}
	sem.Release(1)
}

func TestVSemFacadeErrors(t *testing.T) {
	sem := vsem.New(1)
	if err := sem.Acquire(context.Background(), 2); !errors.Is(err, vsem.ErrInvalidWeight) {
		t.Fatalf("Acquire(2) = %v, want ErrInvalidWeight", err)
	}

	if err := sem.Acquire(context.Background(), 1); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
	defer cancel()
	if err := sem.Acquire(ctx, 1); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("Acquire() timeout = %v, want context deadline", err)
	}
	sem.Close()
	sem.Release(1)
	if err := sem.Acquire(context.Background(), 1); !errors.Is(err, vsem.ErrClosed) {
		t.Fatalf("Acquire() after Close() = %v, want ErrClosed", err)
	}
}
