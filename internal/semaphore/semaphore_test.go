package semaphore

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestAcquireTryAcquireRelease(t *testing.T) {
	sem := New(2)

	if sem.Cap() != 2 {
		t.Fatalf("Cap() = %d, want 2", sem.Cap())
	}
	if err := sem.Acquire(context.Background(), 2); err != nil {
		t.Fatalf("Acquire() error = %v", err)
	}
	if sem.InUse() != 2 {
		t.Fatalf("InUse() = %d, want 2", sem.InUse())
	}
	if sem.TryAcquire(1) {
		t.Fatal("TryAcquire() should fail when capacity is full")
	}

	sem.Release(1)
	if sem.InUse() != 1 {
		t.Fatalf("InUse() after Release() = %d, want 1", sem.InUse())
	}
	if !sem.TryAcquire(1) {
		t.Fatal("TryAcquire() should succeed after one permit is released")
	}
	sem.Release(2)
	if sem.InUse() != 0 {
		t.Fatalf("InUse() after final Release() = %d, want 0", sem.InUse())
	}
}

func TestAcquireWaitsInFIFOOrder(t *testing.T) {
	sem := New(2)
	if err := sem.Acquire(context.Background(), 2); err != nil {
		t.Fatal(err)
	}

	first := make(chan error, 1)
	second := make(chan error, 1)
	go func() { first <- sem.Acquire(context.Background(), 2) }()
	waitUntil(t, func() bool { return queueLen(sem) == 1 })
	go func() { second <- sem.Acquire(context.Background(), 1) }()
	waitUntil(t, func() bool { return queueLen(sem) == 2 })

	sem.Release(1)
	assertNoAcquire(t, first)
	assertNoAcquire(t, second)

	sem.Release(1)
	if err := <-first; err != nil {
		t.Fatalf("first acquire error = %v", err)
	}
	assertNoAcquire(t, second)

	sem.Release(2)
	if err := <-second; err != nil {
		t.Fatalf("second acquire error = %v", err)
	}
	sem.Release(1)
}

func TestAcquireContextCancelRemovesWaiterAndNotifiesNext(t *testing.T) {
	sem := New(2)
	if err := sem.Acquire(context.Background(), 2); err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	first := make(chan error, 1)
	second := make(chan error, 1)
	go func() { first <- sem.Acquire(ctx, 2) }()
	waitUntil(t, func() bool { return queueLen(sem) == 1 })
	go func() { second <- sem.Acquire(context.Background(), 1) }()
	waitUntil(t, func() bool { return queueLen(sem) == 2 })

	cancel()
	if err := <-first; !errors.Is(err, context.Canceled) {
		t.Fatalf("first acquire error = %v, want context.Canceled", err)
	}

	sem.Release(1)
	if err := <-second; err != nil {
		t.Fatalf("second acquire error = %v", err)
	}
	sem.Release(1)
}

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

	sem := New(1)
	var nilCtx context.Context
	if err := sem.Acquire(nilCtx, 0); !errors.Is(err, ErrInvalidWeight) {
		t.Fatalf("Acquire(0) = %v, want ErrInvalidWeight", err)
	}
	if err := sem.Acquire(nilCtx, 2); !errors.Is(err, ErrInvalidWeight) {
		t.Fatalf("Acquire(2) = %v, want ErrInvalidWeight", err)
	}
	if sem.TryAcquire(0) || sem.TryAcquire(2) {
		t.Fatal("TryAcquire() should reject invalid weights")
	}
	assertPanic(t, func() { sem.Release(0) })
	assertPanic(t, func() { sem.Release(1) })
}

func assertNoAcquire(t *testing.T, ch <-chan error) {
	t.Helper()
	select {
	case err := <-ch:
		t.Fatalf("unexpected acquire result: %v", err)
	case <-time.After(30 * time.Millisecond):
	}
}

func assertPanic(t *testing.T, fn func()) {
	t.Helper()
	defer func() {
		if recover() == nil {
			t.Fatal("expected panic")
		}
	}()
	fn()
}

func waitUntil(t *testing.T, fn func() bool) {
	t.Helper()
	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		if fn() {
			return
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatal("condition not met before deadline")
}

func queueLen(sem *Semaphore) int {
	sem.mux.Lock()
	defer sem.mux.Unlock()
	return len(sem.queues)
}
