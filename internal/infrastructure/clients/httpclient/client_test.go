package httpclient

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/sync/errgroup"

	"github.com/jbenzshawel/playlist-generator/internal/infrastructure/clients/httpclient/internal/ratelimit"
)

func TestRetryingClient_Do_RateLimited(t *testing.T) {
	t.Parallel()

	const (
		reqLimit      = 7
		retryAfterDur = 2
	)

	reqCount := atomic.Int32{}
	rateLimitCount := atomic.Int32{}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if reqCount.Load() == reqLimit {
			rateLimitCount.Add(1)

			sync.OnceFunc(func() {
				select {
				case <-t.Context().Done():
				case <-time.After(time.Duration(retryAfterDur) * time.Second):
					reqCount.Swap(0)
				}
			})()

			w.Header().Set("Retry-After", strconv.Itoa(retryAfterDur))
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}

		reqCount.Add(1)
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	rc := NewRetryingClient(Config{})

	g, gCtx := errgroup.WithContext(t.Context())

	req, err := http.NewRequestWithContext(gCtx, "GET", ts.URL, nil)
	require.NoError(t, err)

	numRequests := 15
	statusCodes := make([]int, numRequests)
	for idx := range numRequests {
		g.Go(func() error {
			resp, gErr := rc.Do(req)
			if gErr != nil {
				return gErr
			}

			statusCodes[idx] = resp.StatusCode
			return nil
		})
	}

	require.NoError(t, g.Wait())
	for _, statusCode := range statusCodes {
		require.Equal(t, http.StatusOK, statusCode)
	}

	assert.Greater(t, rateLimitCount.Load(), int32(reqLimit))
}

func TestRetryingClient_Do_ClientLimiter_RateLimited(t *testing.T) {
	t.Parallel()

	const (
		reqLimit      = 7
		retryAfterDur = 2
		batchSize     = 4
	)

	reqCount := atomic.Int32{}
	rateLimitCount := atomic.Int32{}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if reqCount.Load() == reqLimit {
			rateLimitCount.Add(1)

			sync.OnceFunc(func() {
				select {
				case <-t.Context().Done():
				case <-time.After(time.Duration(retryAfterDur) * time.Second):
					reqCount.Swap(0)
				}
			})()

			w.Header().Set("Retry-After", strconv.Itoa(retryAfterDur))
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}

		reqCount.Add(1)
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	rc := NewRetryingClient(Config{
		LimitNumRequests: 5,
		LimitWindow:      batchSize,
	})

	g, gCtx := errgroup.WithContext(t.Context())
	g.SetLimit(batchSize)

	req, err := http.NewRequestWithContext(gCtx, "GET", ts.URL, nil)
	require.NoError(t, err)

	numRequests := 15
	statusCodes := make([]int, numRequests)
	for idx := range numRequests {
		g.Go(func() error {
			resp, gErr := rc.Do(req)
			if gErr != nil {
				return gErr
			}

			statusCodes[idx] = resp.StatusCode
			return nil
		})
	}

	require.NoError(t, g.Wait())
	for _, statusCode := range statusCodes {
		require.Equal(t, http.StatusOK, statusCode)
	}

	assert.LessOrEqual(t, rateLimitCount.Load(), int32(batchSize))
}

func TestRetryingClient_Do_ClientLimiter(t *testing.T) {
	t.Parallel()

	const (
		windowSize  = 2
		maxRequests = 5
		batchSize   = 4
	)

	counter := ratelimit.NewSlidingWindowCounter(windowSize)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Less(t, counter.Count(), int64(maxRequests+1), "handler calls greater than max requests")
		// Include a sleep so there is some backup of workers in the error group
		sleep(t.Context(), 15*time.Millisecond)
		counter.Increment()
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	rc := NewRetryingClient(Config{
		LimitNumRequests: maxRequests,
		LimitWindow:      windowSize,
		LimitBatchSize:   batchSize,
	})

	g, gCtx := errgroup.WithContext(t.Context())
	g.SetLimit(batchSize)

	req, err := http.NewRequestWithContext(gCtx, "GET", ts.URL, nil)
	require.NoError(t, err)

	numRequests := 20
	statusCodes := make([]int, numRequests)
	for idx := range numRequests {
		g.Go(func() error {
			resp, gErr := rc.Do(req)
			if gErr != nil {
				return gErr
			}

			statusCodes[idx] = resp.StatusCode
			return nil
		})
	}

	require.NoError(t, g.Wait())
	for _, statusCode := range statusCodes {
		require.Equal(t, http.StatusOK, statusCode)
	}
}

func sleep(ctx context.Context, duration time.Duration) {
	select {
	case <-ctx.Done():
	case <-time.After(duration):
	}
}
