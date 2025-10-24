package async

import (
	"context"
	"sync"
)

// ParallelFor executes a loop body in parallel using a fixed number of goroutines.
// Now accepts a context and will stop early if the context is canceled.
//
// - ctx: Cancellation/timeout context controlling the loop lifetime.
// - total: The total number of iterations (e.g., 0 to total-1).
// - numWorkers: The number of goroutines to use.
// - loopBody: The function to call for each iteration. It receives the loop index 'i'.
func ParallelFor(ctx context.Context, total int, numWorkers int, loopBody func(i int)) {
	var wg sync.WaitGroup

	// Use a buffered channel to queue up jobs.
	jobs := make(chan int, numWorkers)

	// Start the workers
	for w := 0; w < numWorkers; w++ {
		go func() {
			for idx := range jobs {
				// Ensure every dequeued job decrements the WaitGroup exactly once.
				func() {
					defer wg.Done()
					// If context is already canceled, skip processing but still mark Done.
					if ctx.Err() != nil {
						return
					}
					loopBody(idx)
				}()
			}
		}()
	}

	// Dispatch jobs, honoring context cancellation.
	for idx := 0; idx < total; idx++ {
		// Increment before enqueue to avoid race where worker Done happens before Add.
		wg.Add(1)
		select {
		case <-ctx.Done():
			// Compensate for the Add above since we didn't enqueue this job.
			wg.Done()
			close(jobs)
			wg.Wait()
			return
		case jobs <- idx:
		}
	}
	close(jobs)

	// Wait for all enqueued jobs to finish
	wg.Wait()
}
