// Package ratelimit provides a way to handle rate limits client side.
//
// This package is intended to be used in combination with a http client
// to coordinate concurrent http requests against a rate limited service.

package ratelimit

import (
	"context"
	"log/slog"
	"math/rand/v2"

	"sync"
	"time"
)

const (
	bucketLimitWaitDuration = 3 * time.Second
)

type RateLimit struct {
	lock sync.RWMutex

	limited           bool
	timeAfterDuration time.Duration

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
// keep track of limits configured with SetTimeAfter. To also configure
// client side limiter include WithClientLimits
func New(opts ...RetryLimitOption) *RateLimit {
	cfg := &config{}
	for _, opt := range opts {
		opt(cfg)
	}

	rl := &RateLimit{
		window:      NewSlidingWindowCounter(cfg.window),
		maxRequests: cfg.maxReq,
		batchSize:   cfg.batchSize,
	}

	return rl
}

func (r *RateLimit) Limited() bool {
	r.lock.RLock()
	defer r.lock.RUnlock()

	return r.limited
}

func (r *RateLimit) clearTimeAfter() {
	r.lock.Lock()
	defer r.lock.Unlock()

	r.limited = false
	r.timeAfterDuration = 0

	slog.Debug("client circuit closed")
}

func (r *RateLimit) WaitTimeAfter(ctx context.Context) {
	select {
	case <-ctx.Done():
	case <-time.After(r.getTimeAfterDuration()):
	}
}

func (r *RateLimit) SetTimeAfter(ctx context.Context, d time.Duration) {
	r.lock.Lock()
	defer r.lock.Unlock()

	r.limited = true
	r.timeAfterDuration = d

	go func() {
		select {
		case <-ctx.Done():
		case <-time.After(d):
			r.clearTimeAfter()
		}
	}()

	slog.Debug("client circuit open")
}

func (r *RateLimit) getTimeAfterDuration() time.Duration {
	// include a jitter to prevent subsequent requests from
	// running at the same time and getting rate limited again
	jitter := int64(rand.IntN(250))
	return time.Duration(int64(r.timeAfterDuration) + jitter)
}

func (r *RateLimit) Increment(ctx context.Context) {
	r.window.Increment()

	count := r.window.Count() + r.batchSize
	if r.maxRequests > 0 && count > r.maxRequests {
		slog.Warn("bucket limit reached max requests")
		r.SetTimeAfter(ctx, bucketLimitWaitDuration)
	}

}
