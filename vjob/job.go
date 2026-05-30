package vjob

import (
	"context"

	jobimpl "github.com/imajinyun/go-knifer/internal/job"
)

var (
	// ErrNilJob indicates that a nil Sliceable job was passed to a runner.
	ErrNilJob = jobimpl.ErrNilJob
	// ErrInvalidRange indicates that a Run call received an invalid half-open range.
	ErrInvalidRange = jobimpl.ErrInvalidRange
)

// Merge is called serially by the scheduler after a shard succeeds.
type Merge = jobimpl.Merge

// Sliceable describes work that can be split by half-open index ranges.
type Sliceable = jobimpl.Sliceable

// Options controls scheduling behavior. The zero value is valid.
type Options = jobimpl.Options

// Slice adapts index ranges to the Sliceable interface.
type Slice = jobimpl.Slice

// Batch adapts a typed slice to the Sliceable interface.
type Batch[T any] struct {
	Options
	inner *jobimpl.Batch[T]
}

// Run executes job with the default Options.
func Run(ctx context.Context, job Sliceable) error { return jobimpl.Run(ctx, job) }

// RunWith executes job with explicit scheduling options.
func RunWith(ctx context.Context, job Sliceable, opts Options) error {
	return jobimpl.RunWith(ctx, job, opts)
}

// NewSlice creates a range-based job.
func NewSlice(run func(context.Context, int, int) (Merge, error), length int) *Slice {
	return jobimpl.NewSlice(run, length)
}

// NewSliceSingle creates a job that processes one index per shard.
func NewSliceSingle(run func(context.Context, int) (Merge, error), length int) *Slice {
	return jobimpl.NewSliceSingle(run, length)
}

// NewBatch creates a typed slice job.
func NewBatch[T any](run func(context.Context, []T) (Merge, error), vals []T) *Batch[T] {
	return wrapBatch(jobimpl.NewBatch(run, vals))
}

// NewBatchSingle creates a typed slice job that processes one item per shard.
func NewBatchSingle[T any](run func(context.Context, T) (Merge, error), vals []T) *Batch[T] {
	return wrapBatch(jobimpl.NewBatchSingle(run, vals))
}

// NewMap creates a single-item job over map keys.
func NewMap(run any, m any) *Slice { return jobimpl.NewMap(run, m) }

// NewMapKeys creates a single-item job over typed map keys.
func NewMapKeys[K comparable, V any](run func(context.Context, K) (Merge, error), m map[K]V) *Batch[K] {
	return wrapBatch(jobimpl.NewMapKeys(run, m))
}

// WithBatchSize sets the number of items passed to each Run call.
func (b *Batch[T]) WithBatchSize(size int) *Batch[T] {
	b.ensureInner().WithBatchSize(size)
	b.Options.BatchSize = size
	return b
}

// WithMaxConcurrency sets the maximum number of concurrent shards.
func (b *Batch[T]) WithMaxConcurrency(maxConcurrency int) *Batch[T] {
	b.ensureInner().WithMaxConcurrency(maxConcurrency)
	b.Options.MaxConcurrency = maxConcurrency
	return b
}

// Len returns the number of items to process.
func (b *Batch[T]) Len() int { return b.ensureInner().Len() }

// Run executes the typed worker for items in [start, end).
func (b *Batch[T]) Run(ctx context.Context, start, end int) (Merge, error) {
	return b.ensureInner().Run(ctx, start, end)
}

func wrapBatch[T any](inner *jobimpl.Batch[T]) *Batch[T] {
	if inner == nil {
		return &Batch[T]{}
	}
	return &Batch[T]{Options: inner.Options, inner: inner}
}

func (b *Batch[T]) ensureInner() *jobimpl.Batch[T] {
	if b.inner == nil {
		b.inner = jobimpl.NewBatch[T](nil, nil)
		b.inner.Options = b.Options
	}
	return b.inner
}
