package semaphore

import (
	"context"
	"errors"
	"testing"

	knifer "github.com/imajinyun/knifer-go"
)

func TestCloseWakesWaitersAndRejectsAcquire(t *testing.T) {
	sem := New(1)
	if err := sem.Acquire(context.Background(), 1); err != nil {
		t.Fatal(err)
	}

	waited := make(chan error, 1)
	go func() { waited <- sem.Acquire(context.Background(), 1) }()
	waitUntil(t, func() bool { return queueLen(sem) == 1 })

	sem.Close()
	if err := <-waited; !errors.Is(err, ErrClosed) {
		t.Fatalf("waited acquire error = %v, want ErrClosed", err)
	}
	if err := sem.Acquire(context.Background(), 1); !errors.Is(err, ErrClosed) {
		t.Fatalf("Acquire() after Close() = %v, want ErrClosed", err)
	}
	if sem.TryAcquire(1) {
		t.Fatal("TryAcquire() after Close() should fail")
	}
	sem.Release(1)
}

func TestInvalidInputsAndPanics(t *testing.T) {
	assertPanic(t, func() { New(0) })
	if sem, err := NewE(0); err == nil || sem != nil {
		t.Fatalf("NewE(0) = %v, %v, want nil + error", sem, err)
	} else if !errors.Is(err, ErrInvalidCapacity) {
		t.Fatalf("NewE(0) error = %v, want ErrInvalidCapacity", err)
	} else if !errors.Is(err, knifer.ErrCodeInvalidInput) {
		t.Fatalf("NewE(0) error = %v, want ErrCodeInvalidInput", err)
	}

	sem := New(1)
	var nilCtx context.Context
	if err := sem.Acquire(nilCtx, 0); !errors.Is(err, ErrInvalidWeight) {
		t.Fatalf("Acquire(0) = %v, want ErrInvalidWeight", err)
	} else if !errors.Is(err, knifer.ErrCodeInvalidInput) {
		t.Fatalf("Acquire(0) = %v, want ErrCodeInvalidInput", err)
	}
	if err := sem.Acquire(nilCtx, 2); !errors.Is(err, ErrInvalidWeight) {
		t.Fatalf("Acquire(2) = %v, want ErrInvalidWeight", err)
	}
	if code, ok := knifer.CodeOf(ErrReleaseTooMany); !ok || code != knifer.ErrCodeInvalidInput {
		t.Fatalf("CodeOf(ErrReleaseTooMany) = %q, %v; want invalid input", code, ok)
	}
	if code, ok := knifer.CodeOf(ErrClosed); !ok || code != knifer.ErrCodeUnsupported {
		t.Fatalf("CodeOf(ErrClosed) = %q, %v; want unsupported", code, ok)
	}
	if sem.TryAcquire(0) || sem.TryAcquire(2) {
		t.Fatal("TryAcquire() should reject invalid weights")
	}
	if err := sem.ReleaseE(0); !errors.Is(err, ErrInvalidWeight) {
		t.Fatalf("ReleaseE(0) = %v, want ErrInvalidWeight", err)
	}
	assertPanic(t, func() { sem.Release(0) })
	if err := sem.ReleaseE(1); !errors.Is(err, ErrReleaseTooMany) {
		t.Fatalf("ReleaseE(1) = %v, want ErrReleaseTooMany", err)
	}
	assertPanic(t, func() { sem.Release(1) })
}
