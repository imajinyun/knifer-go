package job

import (
	"context"
	"errors"
)

var (
	// ErrNilJob indicates that a nil Sliceable job was passed to a runner.
	ErrNilJob = errors.New("job is nil")
	// ErrInvalidRange indicates that a Run call received an invalid half-open range.
	ErrInvalidRange = errors.New("job: invalid range")
)

const singleConcurrency = 32

// Merge is called serially by the scheduler after a shard succeeds.
type Merge func() error

// Sliceable describes work that can be split by half-open index ranges.
// Implementations only need to define how data is split and how one shard runs.
type Sliceable interface {
	Len() int
	Run(ctx context.Context, start, end int) (Merge, error)
}

// Options controls scheduling behavior. The zero value is valid.
type Options struct {
	// BatchSize is the number of items per shard. Values <= 0 use Len().
	BatchSize int
	// MaxConcurrency is the maximum number of concurrent shards. Values <= 0 use 1.
	MaxConcurrency int
}

func (o Options) normalized(length int) Options {
	if o.BatchSize <= 0 {
		o.BatchSize = length
	}
	if o.MaxConcurrency <= 0 {
		o.MaxConcurrency = 1
	}
	return o
}

func chunks(length, batchSize int) int {
	if length <= 0 || batchSize <= 0 {
		return 0
	}
	return (length + batchSize - 1) / batchSize
}
