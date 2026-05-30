// Package semaphore provides a weighted, context-aware counting semaphore.
package semaphore

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
)

var (
	ErrInvalidWeight  = errors.New("semaphore: invalid weight")
	ErrReleaseTooMany = errors.New("semaphore: release more than acquired")
	ErrClosed         = errors.New("semaphore: closed")
)

// Semaphore is a weighted, cancellable, and closeable counting semaphore.
// The zero value is not ready to use; construct it with New. It is safe for concurrent use.
type Semaphore struct {
	cap int64
	cur atomic.Int64 // currently acquired weight
	mux sync.Mutex

	queues []*waiter // FIFO wait queue for fairness
	closed bool
}

type waiter struct {
	n     int64
	ready chan struct{} // closed when the waiter is woken
	err   error         // nil means permits were acquired; non-nil means waiting failed
}

// New creates a semaphore with capacity permits. Capacity must be greater than 0.
func New(cap int) *Semaphore {
	if cap <= 0 {
		panic(fmt.Sprintf("semaphore: cap must be > 0, got %d", cap))
	}
	return &Semaphore{cap: int64(cap)}
}

// Acquire obtains n permits, blocking until success, context cancellation, or semaphore close.
func (s *Semaphore) Acquire(ctx context.Context, n int) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if n <= 0 || int64(n) > s.cap {
		return ErrInvalidWeight
	}
	weight := int64(n)

	s.mux.Lock()
	if s.closed {
		s.mux.Unlock()
		return ErrClosed
	}
	// Fast path: no waiters and enough capacity.
	if len(s.queues) == 0 && s.cur.Load()+weight <= s.cap {
		s.cur.Add(weight)
		s.mux.Unlock()
		return nil
	}
	// Slow path: enqueue and wait.
	w := &waiter{n: weight, ready: make(chan struct{})}
	s.queues = append(s.queues, w)
	s.mux.Unlock()

	select {
	case <-w.ready:
		// Woken by Release or Close. If permits have already been granted,
		// return success even if Close happens later, otherwise the permits
		// would be counted in cur without a caller that can release them.
		s.mux.Lock()
		defer s.mux.Unlock()
		return w.err
	case <-ctx.Done():
		// Cancellation: remove from the queue; if already woken, compensate by releasing.
		s.mux.Lock()
		select {
		case <-w.ready:
			err := w.err
			if err != nil {
				s.mux.Unlock()
				return err
			}
			// Permits were granted but the caller was canceled; release immediately
			// without blocking on the canceled context.
			s.mux.Unlock()
			s.Release(int(weight))
			return ctx.Err()
		default:
			s.removeWaiterLocked(w)
			s.notifyLocked()
			s.mux.Unlock()
			return ctx.Err()
		}
	}
}

// TryAcquire attempts to obtain n permits without blocking.
func (s *Semaphore) TryAcquire(n int) bool {
	if n <= 0 || int64(n) > s.cap {
		return false
	}
	weight := int64(n)

	s.mux.Lock()
	defer s.mux.Unlock()
	if s.closed || len(s.queues) > 0 {
		return false
	}
	if s.cur.Load()+weight <= s.cap {
		s.cur.Add(weight)
		return true
	}
	return false
}

// Release releases n permits and wakes waiters in FIFO order.
func (s *Semaphore) Release(n int) {
	if n <= 0 {
		panic(ErrInvalidWeight)
	}
	weight := int64(n)

	s.mux.Lock()
	defer s.mux.Unlock()

	if s.cur.Load() < weight {
		panic(ErrReleaseTooMany)
	}
	s.cur.Add(-weight)
	s.notifyLocked()
}

// InUse returns the number of currently acquired permits.
func (s *Semaphore) InUse() int { return int(s.cur.Load()) }

// Cap returns the total capacity.
func (s *Semaphore) Cap() int { return int(s.cap) }

// Close closes the semaphore and wakes all waiters with ErrClosed.
// Already acquired permits are not forcibly reclaimed; callers still need to Release them.
func (s *Semaphore) Close() {
	s.mux.Lock()
	defer s.mux.Unlock()
	if s.closed {
		return
	}
	s.closed = true
	for _, w := range s.queues {
		w.err = ErrClosed
		close(w.ready)
	}
	s.queues = nil
}

func (s *Semaphore) notifyLocked() {
	for len(s.queues) > 0 {
		w := s.queues[0]
		if s.cur.Load()+w.n > s.cap {
			return // Keep FIFO fairness: do not let later waiters jump ahead.
		}
		s.cur.Add(w.n)
		s.queues = s.queues[1:]
		w.err = nil
		close(w.ready)
	}
}

func (s *Semaphore) removeWaiterLocked(target *waiter) {
	for i, w := range s.queues {
		if w == target {
			s.queues = append(s.queues[:i], s.queues[i+1:]...)
			return
		}
	}
}
