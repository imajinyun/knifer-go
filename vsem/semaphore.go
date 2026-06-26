package vsem

import semimpl "github.com/imajinyun/knifer-go/internal/semaphore"

var (
	// ErrInvalidCapacity indicates an invalid semaphore capacity.
	ErrInvalidCapacity = semimpl.ErrInvalidCapacity
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

// NewE creates a semaphore with capacity permits and returns an error for invalid capacity.
func NewE(capacity int) (*Semaphore, error) { return semimpl.NewE(capacity) }
