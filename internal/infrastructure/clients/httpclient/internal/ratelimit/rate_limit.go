// Package ratelimit provides a way to handle rate limits client side.
//
// This package is intended to be used in combination with a http client
// to coordinate concurrent http requests against a rate limited service.
package ratelimit

import (
	"context"
	"log/slog"
	"sync"
	"sync/atomic"
	"time"
)

const (
	bucketLimitWaitDuration = 3 * time.Second
)

type RateLimit struct {
	mu        sync.Mutex
	limited   atomic.Bool
	limitDone chan struct{}

	maxRequests int64
	batchSize   int64
	window      SlidingWindowCounter
}

type config struct {
	window    int64
	maxReq    int64
	batchSize int64
}

type RetryLimitOption func(*config)

// WithClientLimits can be used to provide a client side rate limit. When the
// max requests is hit within a sliding window the RateLimit is limited for the
// bucketLimitWaitDuration
func WithClientLimits(window, maxReq, batchSize int) RetryLimitOption {
	return func(c *config) {
		c.window = int64(window)
		c.maxReq = int64(maxReq)
		c.batchSize = int64(batchSize)
	}
}

// New returns a RateLimit. When no options configured instance will only
// keep track of limits configured with SetLimited. To also configure
// client side limiter include WithClientLimits
func New(opts ...RetryLimitOption) *RateLimit {
	cfg := &config{}
	for _, opt := range opts {
		opt(cfg)
	}

	// Create a channel and immediately close it to set the initial not limited
	// state to "done"
	doneChan := make(chan struct{})
	close(doneChan)

	rl := &RateLimit{
		window:      NewSlidingWindowCounter(cfg.window),
		maxRequests: cfg.maxReq,
		batchSize:   cfg.batchSize,
		limitDone:   doneChan,
	}

	return rl
}

// Limited returns true when requests should be delayed/queued  due to a rate limit
func (r *RateLimit) Limited() bool {
	return r.limited.Load()
}

// SetLimited will configure the RateLimit to be Limited for the configured
// duration.
func (r *RateLimit) SetLimited(d time.Duration) {
	if r.limited.Load() {
		return
	}

	r.mu.Lock()
	// Check after lock in case value changes after lock acquired
	if r.limited.Load() {
		r.mu.Unlock()
		return
	}
	r.limited.Swap(true)
	doneChan := make(chan struct{})
	r.limitDone = doneChan
	r.mu.Unlock()

	slog.Debug("client circuit open")

	go func() {
		// We do NOT select on ctx.Done() here. The clearing of a global
		// rate limit should not be tied to the context of the request that
		// triggered it since cancelling thr original request's context could
		// lead to the doneChan never closing and circuit always being open.
		<-time.After(d)

		r.limited.Swap(false)
		close(doneChan)

		slog.Debug("client circuit closed")
	}()

}

// WaitTimeAfter can be called when Limited returns true to wait until
// a rate limit time after has lapsed.
func (r *RateLimit) WaitTimeAfter(ctx context.Context) {
	start := time.Now()
	defer func() {
		slog.Debug("closed circuit wait time:", slog.Any("dur", time.Since(start).Seconds()))
	}()

	if !r.limited.Load() {
		return
	}

	r.mu.Lock()
	// Check after lock in case value changes after lock acquired
	if !r.limited.Load() {
		r.mu.Unlock()
		return
	}
	doneChan := r.limitDone
	r.mu.Unlock()

	select {
	case <-ctx.Done():
		return
	case <-doneChan: // This will unblock when close(doneChan) is called
		return
	}
}

// Increment increments the number of requests in the configured sliding window.
// If the count reaches the window's max requests, the RateLimit is Limited for
// the fixed bucketLimitWaitDuration (TODO: make this configurable?).
//
// Note: if the RateLimit was not configured WithClientLimits this is a noop.
func (r *RateLimit) Increment() {
	if r.limited.Load() {
		return
	}

	r.window.Increment()

	// Include batchSize in count to better account for concurrent requests.
	count := r.window.Count() + r.batchSize
	if r.maxRequests > 0 && count > r.maxRequests {
		slog.Warn("bucket limit reached max requests")
		r.SetLimited(bucketLimitWaitDuration)
	}
}

// Count returns the current request count in the sliding window
func (r *RateLimit) Count() int64 {
	return r.window.Count()
}
