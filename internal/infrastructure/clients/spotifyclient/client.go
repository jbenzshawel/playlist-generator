package spotifyclient

import (
	"context"
	"fmt"
	"net/url"

	"github.com/jbenzshawel/playlist-generator/internal/app/playlists/spotify"
	"github.com/jbenzshawel/playlist-generator/internal/infrastructure/clients/httpclient"
	"github.com/jbenzshawel/playlist-generator/internal/infrastructure/clients/httpclient/auth"
	"github.com/jbenzshawel/playlist-generator/internal/infrastructure/clients/httpclient/decode"
)

type Config struct {
	BaseURL *url.URL
	Auth    auth.Config
}

type client struct {
	httpclient.Client
}

func New(cfg Config) *client {
	return &client{
		Client: httpclient.NewRetryingClient(httpclient.Config{
			BaseURL: cfg.BaseURL,
			Auth:    &cfg.Auth,
		}),
	}
}

func (c *client) SearchTrack(ctx context.Context, artist, track, album string) (spotify.SearchTrackResponse, error) {
	query := fmt.Sprintf("%s%%20artist:%s", url.QueryEscape(track), url.QueryEscape(artist))
	if album != "" {
		query += fmt.Sprintf("%%20album:%s", url.QueryEscape(album))
	}

	resp, err := c.Get(ctx, "/search", httpclient.WithQuery(map[string]string{
		"q":    query,
		"type": "track",
	}))
	if err != nil {
		return spotify.SearchTrackResponse{}, err
	}

	defer resp.Body.Close()

	collection, err := decode.JSON[spotify.SearchTrackResponse](resp)
	if err != nil {
		return spotify.SearchTrackResponse{}, err
	}

	return collection, nil
}
