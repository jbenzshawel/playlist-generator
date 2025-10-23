package spotifyclient

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/jbenzshawel/playlist-generator/internal/app/playlists/spotify"
	"github.com/jbenzshawel/playlist-generator/internal/infrastructure/clients/httpclient"
	"github.com/jbenzshawel/playlist-generator/internal/infrastructure/clients/httpclient/decode"
)

type Config struct {
	BaseURL *url.URL
	Client  *http.Client
}

type client struct {
	httpclient.Client
}

func New(cfg Config) *client {
	return &client{
		Client: httpclient.NewRetryingClient(httpclient.Config{
			BaseURL: cfg.BaseURL,
			Client:  cfg.Client,
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

func (c *client) CurrentUser(ctx context.Context) (spotify.User, error) {
	resp, err := c.Get(ctx, "/me")
	if err != nil {
		return spotify.User{}, err
	}

	defer resp.Body.Close()

	user, err := decode.JSON[spotify.User](resp)
	if err != nil {
		return spotify.User{}, err
	}

	return user, nil
}

func (c *client) CreatePlaylist(ctx context.Context, userID string, request spotify.CreatePlaylistRequest) (spotify.SimplePlaylist, error) {
	resp, err := c.Post(ctx, fmt.Sprintf("/users/%s/platlists", userID), httpclient.WithJSONBody(request))
	if err != nil {
		return spotify.SimplePlaylist{}, err
	}

	defer resp.Body.Close()

	playlist, err := decode.JSON[spotify.SimplePlaylist](resp)
	if err != nil {
		return spotify.SimplePlaylist{}, err
	}

	return playlist, nil
}
