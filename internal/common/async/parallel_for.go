package async

import (
	"context"

	"golang.org/x/sync/errgroup"
)

// ParallelFor executes a loop body in parallel across a pool of workers.
// It stops dispatching new jobs and attempts to stop running jobs
// as soon as the context is canceled or the first error occurs.
func ParallelFor(ctx context.Context, total int, numWorkers int, fn func(ctx context.Context, i int) error) error {
	g, gCtx := errgroup.WithContext(ctx)

	g.SetLimit(numWorkers)

	for idx := 0; idx < total; idx++ {
		g.Go(func() error {
			select {
			case <-gCtx.Done():
				return gCtx.Err()
			default:
				return fn(gCtx, idx)
			}
		})
	}

	err := g.Wait()
	if err != nil {
		return err
	}

	return err
}
