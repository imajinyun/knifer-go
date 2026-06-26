package job

import (
	"context"

	"github.com/imajinyun/knifer-go/internal/errx"
	"github.com/imajinyun/knifer-go/internal/semaphore"
)

// Run executes job with the default Options. By default, the whole job runs as one serial shard.
func Run(ctx context.Context, job Sliceable) error {
	if carrier, ok := job.(OptionCarrier); ok {
		return RunWith(ctx, job, carrier.JobOptions())
	}
	return RunWith(ctx, job, Options{})
}

// RunWith executes job with explicit scheduling options.
func RunWith(ctx context.Context, job Sliceable, opts Options) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if job == nil {
		return ErrNilJob
	}

	length := job.Len()
	if length <= 0 {
		return nil
	}
	if err := ctx.Err(); err != nil {
		return err
	}

	opts = opts.normalized(length)
	chunkCount := chunks(length, opts.BatchSize)
	merges := make([]Merge, chunkCount)

	if opts.MaxConcurrency == 1 || chunkCount == 1 {
		return runSerial(ctx, job, opts, merges)
	}
	return runConcurrent(ctx, job, opts, merges)
}

func runSerial(ctx context.Context, job Sliceable, opts Options, merges []Merge) error {
	errs := errx.NewCollector().WithContext(ctx)
	length := job.Len()
	for idx, start := 0, 0; start < length; idx, start = idx+1, start+opts.BatchSize {
		end := min(start+opts.BatchSize, length)
		if err := errs.Recover(func() error {
			if err := ctx.Err(); err != nil {
				return err
			}
			merge, err := job.Run(ctx, start, end)
			merges[idx] = merge
			return err
		}, "job run %T shard fail, start=%d end=%d", job, start, end); err != nil {
			return err
		}
	}
	return replayMerges(ctx, job, merges)
}

func runConcurrent(ctx context.Context, job Sliceable, opts Options, merges []Merge) error {
	length := job.Len()
	dispatchCtx, cancelDispatch := context.WithCancel(ctx)
	defer cancelDispatch()

	sem := semaphore.New(opts.MaxConcurrency)
	defer sem.Close()

	errs := errx.NewCollector().WithContext(ctx)
	for idx, start := 0, 0; start < length; idx, start = idx+1, start+opts.BatchSize {
		if err := sem.Acquire(dispatchCtx, 1); err != nil {
			if ctx.Err() != nil {
				errs.Collect(ctx.Err())
			}
			break
		}

		idx, start, end := idx, start, min(start+opts.BatchSize, length)
		errs.GoRun(func() error {
			defer sem.Release(1)
			if err := dispatchCtx.Err(); err != nil {
				return err
			}

			merge, err := job.Run(dispatchCtx, start, end)
			if err != nil {
				cancelDispatch()
				return err
			}
			merges[idx] = merge
			return nil
		}, "job run %T shard fail, start=%d end=%d", job, start, end)
	}

	if err := errs.Error(); err != nil {
		return err
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	return replayMerges(ctx, job, merges)
}

func replayMerges(ctx context.Context, job Sliceable, merges []Merge) error {
	errs := errx.NewCollector().WithContext(ctx)
	for idx, merge := range merges {
		if merge == nil {
			continue
		}
		if err := errs.Recover(func() error {
			if err := ctx.Err(); err != nil {
				return err
			}
			return merge()
		}, "job merge %T shard fail, index=%d", job, idx); err != nil {
			return err
		}
	}
	return nil
}
