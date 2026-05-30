package vsem

import semimpl "github.com/imajinyun/go-knifer/internal/semaphore"

var (
	// ErrInvalidWeight indicates an invalid acquire/release weight.
	ErrInvalidWeight = semimpl.ErrInvalidWeight
	// ErrReleaseTooMany indicates more permits were released than acquired.
	ErrReleaseTooMany = semimpl.ErrReleaseTooMany
	// ErrClosed indicates the semaphore has been closed.
	ErrClosed = semimpl.ErrClosed
)

// Semaphore is a weighted, context-aware counting semaphore.
type Semaphore = semimpl.Semaphore

// New creates a semaphore with capacity permits.
func New(capacity int) *Semaphore { return semimpl.New(capacity) }
