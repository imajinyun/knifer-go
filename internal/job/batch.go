package job

import "context"

var _ Sliceable = (*Batch[int])(nil)

// Batch adapts a typed slice to the Sliceable interface.
type Batch[T any] struct {
	Options

	vals []T
	run  func(context.Context, []T) (Merge, error)
}

// NewBatch creates a typed slice job.
func NewBatch[T any](run func(context.Context, []T) (Merge, error), vals []T) *Batch[T] {
	return &Batch[T]{vals: vals, run: run}
}

// NewBatchSingle creates a typed slice job that processes one item per shard.
func NewBatchSingle[T any](run func(context.Context, T) (Merge, error), vals []T) *Batch[T] {
	return NewBatch(func(ctx context.Context, batch []T) (Merge, error) {
		if run == nil {
			return nil, nil
		}
		return run(ctx, batch[0])
	}, vals).WithBatchSize(1).WithMaxConcurrency(singleConcurrency)
}

// WithBatchSize sets the number of items passed to each Run call.
func (b *Batch[T]) WithBatchSize(size int) *Batch[T] {
	b.BatchSize = size
	return b
}

// WithMaxConcurrency sets the maximum number of concurrent shards.
func (b *Batch[T]) WithMaxConcurrency(maxConcurrency int) *Batch[T] {
	b.MaxConcurrency = maxConcurrency
	return b
}

// Len returns the number of items to process.
func (b *Batch[T]) Len() int { return len(b.vals) }

// Run executes the typed worker for items in [start, end).
func (b *Batch[T]) Run(ctx context.Context, start, end int) (Merge, error) {
	if b.run == nil {
		return nil, nil
	}
	if start < 0 || end < start || end > len(b.vals) {
		return nil, ErrInvalidRange
	}
	return b.run(ctx, b.vals[start:end])
}
