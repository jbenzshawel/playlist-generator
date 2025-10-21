package iprclient

import (
	"context"
	"github.com/jbenzshawel/playlist-generator/internal/clients/httpclient"
	"net/url"
)

type Config struct {
	BaseURL *url.URL
}

type client struct {
	httpclient.Client
}

func New(cfg Config) *client {
	return &client{
		Client: httpclient.NewRetryingClient(httpclient.Config{
			BaseURL: cfg.BaseURL,
		}),
	}
}

func (c *client) GetSongs(ctx context.Context, date string) (Collection, error) {
	resp, err := c.Get(ctx, "/day", httpclient.WithQuery(map[string]string{
		"format": "json",
		"date":   date,
	}))
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
