package vsem_test

import (
	"context"
	"errors"
	"testing"
	"time"

	knifer "github.com/imajinyun/knifer-go"
	"github.com/imajinyun/knifer-go/vsem"
)

func TestVSemFacade(t *testing.T) {
	sem := vsem.New(1)
	if sem.Cap() != 1 {
		t.Fatalf("Cap() = %d, want 1", sem.Cap())
	}
	if err := sem.Acquire(context.Background(), 1); err != nil {
		t.Fatal(err)
	}
	if sem.Use() != 1 {
		t.Fatalf("Use() = %d, want 1", sem.Use())
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
	if sem, err := vsem.NewE(0); err == nil || sem != nil {
		t.Fatalf("NewE(0) = %v, %v, want nil + error", sem, err)
	} else if !errors.Is(err, vsem.ErrInvalidCapacity) {
		t.Fatalf("NewE(0) error = %v, want ErrInvalidCapacity", err)
	}

	sem := vsem.New(1)
	if err := sem.Acquire(context.Background(), 2); !errors.Is(err, vsem.ErrInvalidWeight) {
		t.Fatalf("Acquire(2) = %v, want ErrInvalidWeight", err)
	} else if !errors.Is(err, knifer.ErrCodeInvalidInput) {
		t.Fatalf("Acquire(2) = %v, want ErrCodeInvalidInput", err)
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
	} else if !errors.Is(err, knifer.ErrCodeUnsupported) {
		t.Fatalf("Acquire() after Close() = %v, want ErrCodeUnsupported", err)
	}
	if code, ok := knifer.CodeOf(vsem.ErrInvalidWeight); !ok || code != knifer.ErrCodeInvalidInput {
		t.Fatalf("CodeOf(ErrInvalidWeight) = %q, %v; want invalid input", code, ok)
	}
	active := vsem.New(1)
	if err := active.ReleaseE(0); !errors.Is(err, vsem.ErrInvalidWeight) {
		t.Fatalf("ReleaseE(0) = %v, want ErrInvalidWeight", err)
	}
	if err := active.Acquire(context.Background(), 1); err != nil {
		t.Fatal(err)
	}
	if err := active.ReleaseE(1); err != nil {
		t.Fatalf("ReleaseE(1) = %v, want nil", err)
	}
	if err := active.ReleaseE(1); !errors.Is(err, vsem.ErrReleaseTooMany) {
		t.Fatalf("ReleaseE(1) second = %v, want ErrReleaseTooMany", err)
	}
}
