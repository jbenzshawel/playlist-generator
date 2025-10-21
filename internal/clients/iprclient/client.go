package iprclient

import (
	"context"
	"fmt"

	"github.com/jbenzshawel/playlist-generator/internal/clients/httpclient"
)

type Config struct {
	BaseURL string
}

type client struct {
	httpclient.Client
	baseURL string
}

func New(cfg Config) *client {
	return &client{
		Client:  httpclient.NewRetryingClient(),
		baseURL: cfg.BaseURL,
	}
}

func (c *client) GetSongs(ctx context.Context, date string) (Collection, error) {
	// TODO: use URL?
	endpoint := fmt.Sprintf("%s/day", c.baseURL)

	req, err := httpclient.NewRequest(ctx, endpoint, httpclient.WithQuery(map[string]string{
		"format": "json",
		"date":   date,
	}))
	if err != nil {
		return Collection{}, err
	}

	resp, err := c.Do(req)
	if err != nil {
		return Collection{}, err
	}

	defer resp.Body.Close()

	collection, err := httpclient.DecodeJSON[Collection](resp)
	if err != nil {
		return Collection{}, err
	}

	return collection, nil
}
