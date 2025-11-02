package httpclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"math"
	"math/rand/v2"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/jbenzshawel/playlist-generator/internal/infrastructure/clients/httpclient/internal/ratelimit"
)

const (
	defaultMaxRetries  = 3
	defaultMinWaitTime = time.Duration(100) * time.Millisecond
	defaultMaxWaitTime = time.Duration(2000) * time.Millisecond
)

type Client interface {
	Get(ctx context.Context, endpoint string, options ...RequestOption) (*http.Response, error)
	Post(ctx context.Context, endpoint string, options ...RequestOption) (*http.Response, error)
	Do(req *http.Request) (*http.Response, error)
}

type retryingClient struct {
	client  *http.Client
	baseURL *url.URL

	maxRetries  int
	minWaitTime time.Duration
	maxWaitTime time.Duration

	rateLimit *ratelimit.RateLimit
}

type Config struct {
	BaseURL *url.URL
	Client  *http.Client

	// LimitWindow optional window, in seconds, for client side limiter
	LimitWindow int
	// LimitNumRequests optional max num requests in a window
	LimitNumRequests int
	// LimitBatchSize should be set if client requests will be batched. Configuring
	// this value takes into account batch size when calculating client size limits.
	LimitBatchSize int
}

// NewRetryingClient creates a retryingClient with default settings.
func NewRetryingClient(cfg Config) *retryingClient {
	c := &retryingClient{
		baseURL:     cfg.BaseURL,
		maxRetries:  defaultMaxRetries,
		minWaitTime: defaultMinWaitTime,
		maxWaitTime: defaultMaxWaitTime,
	}

	if cfg.Client != nil {
		c.client = cfg.Client
	} else {
		c.client = &http.Client{Timeout: 10 * time.Second}
	}

	if cfg.LimitNumRequests > 0 {
		c.rateLimit = ratelimit.New(ratelimit.WithClientLimits(cfg.LimitWindow, cfg.LimitNumRequests, cfg.LimitBatchSize))
	} else {
		c.rateLimit = ratelimit.New()
	}

	return c
}

type RequestConfig struct {
	queryParams map[string]string
	jsonBody    any
}

type RequestOption func(*RequestConfig)

func WithQuery(params map[string]string) RequestOption {
	return func(cfg *RequestConfig) {
		cfg.queryParams = params
	}
}

func WithJSONBody(b any) RequestOption {
	return func(cfg *RequestConfig) {
		cfg.jsonBody = b
	}
}

func (c *retryingClient) Get(ctx context.Context, endpoint string, options ...RequestOption) (*http.Response, error) {
	requestURL := c.baseURL.JoinPath(endpoint).String()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL, nil)
	if err != nil {
		return nil, err
	}

	cfg := &RequestConfig{}
	for _, opt := range options {
		opt(cfg)
	}

	q := req.URL.Query()

	for k, v := range cfg.queryParams {
		q.Add(k, v)
	}

	req.URL.RawQuery = q.Encode()

	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (c *retryingClient) Post(ctx context.Context, endpoint string, options ...RequestOption) (*http.Response, error) {
	requestURL := c.baseURL.JoinPath(endpoint).String()

	cfg := &RequestConfig{}
	for _, opt := range options {
		opt(cfg)
	}

	bodyJSON, err := json.Marshal(cfg.jsonBody)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, requestURL, bytes.NewBuffer(bodyJSON))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (c *retryingClient) Do(req *http.Request) (*http.Response, error) {
	for attempt := 0; attempt < c.maxRetries; attempt++ {
		if c.rateLimit.Limited() {
			// if the circuit is open due to being rate limited
			// wait until the time after elapsed before continuing
			c.rateLimit.WaitTimeAfter(req.Context())
		}

		clone := req.Clone(req.Context())

		wait := c.defaultWaitStrategy(attempt)

		c.rateLimit.Increment(req.Context())

		resp, err := c.client.Do(clone)
		if err != nil {
			slog.Warn("http request failed with network error",
				slog.Any("error", err),
				slog.Int("attempt", attempt),
			)
			c.sleep(req.Context(), wait, false)
			continue
		}

		if !shouldRetry(resp.StatusCode) {
			return resp, nil
		}

		// close the response if we're retrying
		resp.Body.Close()

		if resp.StatusCode == http.StatusTooManyRequests {
			wait = getRetryAfter(resp, wait)
			slog.Warn("http request failed with too many requests",
				slog.Int("attempt", attempt),
				slog.Int("wait", int(wait.Seconds())),
				slog.Int64("requestCount", c.rateLimit.Count()),
			)
			c.sleep(req.Context(), wait, true)
			continue
		}

		if resp.StatusCode >= http.StatusInternalServerError && resp.StatusCode != http.StatusNotImplemented {
			slog.Warn("http request failed with internal server error",
				slog.Int("attempt", attempt),
				slog.Int("wait", int(wait)),
			)
			c.sleep(req.Context(), wait, false)
			continue
		}

		return resp, nil
	}

	return nil, fmt.Errorf("http request failed after max retries %d", c.maxRetries)
}

func (c *retryingClient) defaultWaitStrategy(attempt int) time.Duration {
	wait := math.Min(float64(c.maxWaitTime), float64(c.minWaitTime)*math.Exp2(float64(attempt)))
	center := time.Duration(wait / 2)

	interval := int(center)
	jitter := rand.IntN(interval)
	return time.Duration(math.Abs(float64(interval + jitter)))
}

func (c *retryingClient) sleep(ctx context.Context, d time.Duration, isRateLimited bool) {
	if isRateLimited {
		c.rateLimit.SetLimited(ctx, d)
	}

	select {
	case <-ctx.Done():
	case <-time.After(d):
	}
}

func getRetryAfter(resp *http.Response, defaultWait time.Duration) time.Duration {
	wait := defaultWait

	if header := resp.Header.Get("Retry-After"); header != "" {
		if secs, err := strconv.Atoi(header); err == nil {
			wait = time.Duration(secs) * time.Second
		} else if t, err := time.Parse(http.TimeFormat, header); err == nil {
			wait = time.Until(t)
			if wait < 0 {
				wait = defaultWait
			}
		}
	}

	return wait
}

func shouldRetry(statusCode int) bool {
	return statusCode == http.StatusTooManyRequests ||
		(statusCode > http.StatusInternalServerError && statusCode != http.StatusNotImplemented)
}
