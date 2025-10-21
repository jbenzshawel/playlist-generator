package httpclient

import (
	"context"
	"net/http"
)

type RequestConfig struct {
	queryParams map[string]string
}

type RequestOption func(*RequestConfig)

func WithQuery(params map[string]string) RequestOption {
	return func(cfg *RequestConfig) {
		cfg.queryParams = params
	}
}

func NewRequest(ctx context.Context, endpoint string, options ...RequestOption) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
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

	return req, nil
}
