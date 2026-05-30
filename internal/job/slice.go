package job

import "context"

var _ Sliceable = (*Slice)(nil)

// Slice adapts index ranges to the Sliceable interface.
type Slice struct {
	Options

	run    func(context.Context, int, int) (Merge, error)
	length int
}

// NewSlice creates a range-based job.
func NewSlice(run func(context.Context, int, int) (Merge, error), length int) *Slice {
	return &Slice{run: run, length: length}
}

// NewSliceSingle creates a job that processes one index per shard.
func NewSliceSingle(run func(context.Context, int) (Merge, error), length int) *Slice {
	return NewSlice(func(ctx context.Context, start, end int) (Merge, error) {
		if run == nil {
			return nil, nil
		}
		return run(ctx, start)
	}, length).WithBatchSize(1).WithMaxConcurrency(singleConcurrency)
}

// WithBatchSize sets the number of indexes passed to each Run call.
func (s *Slice) WithBatchSize(size int) *Slice {
	s.BatchSize = size
	return s
}

// WithMaxConcurrency sets the maximum number of concurrent shards.
func (s *Slice) WithMaxConcurrency(maxConcurrency int) *Slice {
	s.MaxConcurrency = maxConcurrency
	return s
}

// Len returns the number of indexes to process.
func (s *Slice) Len() int { return s.length }

// Run executes the range-based worker for [start, end).
func (s *Slice) Run(ctx context.Context, start, end int) (Merge, error) {
	if s.run == nil {
		return nil, nil
	}
	if start < 0 || end < start || end > s.length {
		return nil, ErrInvalidRange
	}
	return s.run(ctx, start, end)
}
