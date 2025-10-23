package httpclient

import (
	"context"
	"fmt"
	"log/slog"
	"math"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"time"

	"github.com/jbenzshawel/playlist-generator/internal/infrastructure/clients/httpclient/auth"
)

const (
	defaultMaxRetries  = 3
	defaultMinWaitTime = time.Duration(100) * time.Millisecond
	defaultMaxWaitTime = time.Duration(2000) * time.Millisecond
)

type Client interface {
	Get(ctx context.Context, endpoint string, options ...RequestOption) (*http.Response, error)
	Do(req *http.Request) (*http.Response, error)
}

type retryingClient struct {
	client  *http.Client
	baseURL *url.URL
	auth    *auth.TokenGetter

	lock sync.Mutex
	rnd  *rand.Rand

	maxRetries  int
	minWaitTime time.Duration
	maxWaitTime time.Duration
}

type Config struct {
	BaseURL *url.URL
	Auth    *auth.Config
}

// NewRetryingClient creates a retryingClient with default settings.
func NewRetryingClient(cfg Config) *retryingClient {
	c := &retryingClient{
		client:      &http.Client{Timeout: 10 * time.Second},
		rnd:         rand.New(rand.NewSource(time.Now().UnixNano())),
		baseURL:     cfg.BaseURL,
		maxRetries:  defaultMaxRetries,
		minWaitTime: defaultMinWaitTime,
		maxWaitTime: defaultMaxWaitTime,
	}

	if cfg.Auth != nil {
		c.auth = &auth.TokenGetter{Cfg: *cfg.Auth}
	}

	return c
}

type RequestConfig struct {
	queryParams map[string]string
}

type RequestOption func(*RequestConfig)

func WithQuery(params map[string]string) RequestOption {
	return func(cfg *RequestConfig) {
		cfg.queryParams = params
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

	if c.auth != nil {
		req, err = c.addAuthHeader(ctx, req)
		if err != nil {
			return nil, err
		}
	}

	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (c *retryingClient) addAuthHeader(ctx context.Context, req *http.Request) (*http.Request, error) {
	token, err := c.auth.GetToken(ctx, c)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))
	return req, nil
}

func (c *retryingClient) Do(req *http.Request) (*http.Response, error) {
	for attempt := 0; attempt < c.maxRetries; attempt++ {
		clone := req.Clone(req.Context())

		resp, err := c.client.Do(clone)

		wait := c.defaultWaitStrategy(attempt)

		if err != nil {
			slog.Warn("http request failed with network error",
				slog.Any("error", err),
				slog.Int("attempt", attempt),
			)
			time.Sleep(wait)
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
			)
			time.Sleep(wait)
			continue
		}

		if resp.StatusCode >= http.StatusInternalServerError && resp.StatusCode != http.StatusNotImplemented {
			slog.Warn("http request failed with internal server error",
				slog.Int("attempt", attempt),
			)
			time.Sleep(wait)
			continue
		}

		return resp, nil
	}

	return nil, fmt.Errorf("http request failed after max retries %d", c.maxRetries)
}

func (c *retryingClient) defaultWaitStrategy(attempt int) time.Duration {
	c.lock.Lock()
	defer c.lock.Unlock()

	wait := math.Min(float64(c.maxWaitTime), float64(c.minWaitTime)*math.Exp2(float64(attempt)))
	center := time.Duration(wait / 2)

	interval := int64(center)
	jitter := c.rnd.Int63n(interval)
	return time.Duration(math.Abs(float64(interval + jitter)))
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
