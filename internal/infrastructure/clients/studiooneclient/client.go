package studiooneclient

import (
	"context"
	"net/url"

	"github.com/jbenzshawel/playlist-generator/internal/app/commands/sources/studioone/models"
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

func (c *client) GetSongs(ctx context.Context, date string) (models.Collection, error) {
	resp, err := c.Get(ctx, "/day", httpclient.WithQuery(map[string]string{
		"format": "json",
		"date":   date,
	}))
	if err != nil {
		return models.Collection{}, err
	}

	defer resp.Body.Close()

	collection, err := decode.JSON[models.Collection](resp)
	if err != nil {
		return models.Collection{}, err
	}

	return collection, nil
}
