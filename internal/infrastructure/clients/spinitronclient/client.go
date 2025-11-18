package spinitronclient

import (
	"context"
	"net/url"

	"github.com/jbenzshawel/playlist-generator/internal/infrastructure/clients/httpclient"
	"github.com/jbenzshawel/playlist-generator/internal/infrastructure/clients/httpclient/decode"
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

func (c *client) ScrapePlaylist(ctx context.Context, source string) ([]byte, error) {
	resp, err := c.Get(ctx, source)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	res, err := decode.Bytes(resp)
	if err != nil {
		return nil, err
	}
	return res, nil
}
