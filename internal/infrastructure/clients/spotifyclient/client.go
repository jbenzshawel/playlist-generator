package spotifyclient

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/jbenzshawel/playlist-generator/internal/app/commands/playlists/spotify/models"
	"github.com/jbenzshawel/playlist-generator/internal/infrastructure/clients/httpclient"
	"github.com/jbenzshawel/playlist-generator/internal/infrastructure/clients/httpclient/decode"
)

type Config struct {
	BaseURL *url.URL
	Client  *http.Client
}

type Client struct {
	httpclient.Client
}

func New(cfg Config) *Client {
	return &Client{
		Client: httpclient.NewRetryingClient(httpclient.Config{
			BaseURL:          cfg.BaseURL,
			Client:           cfg.Client,
			LimitWindow:      30,
			LimitNumRequests: 175,
			LimitBatchSize:   10,
		}),
	}
}

func (c *Client) SearchTrack(ctx context.Context, artist, track, album string) (models.SearchTrackResponse, error) {
	query := fmt.Sprintf("track:%s artist:%s", track, artist)
	if album != "" {
		query += fmt.Sprintf(" album:%s", album)
	}

	resp, err := c.Get(ctx, "/search", httpclient.WithQuery(map[string]string{
		"q":    url.QueryEscape(query),
		"type": "track",
	}))
	if err != nil {
		return models.SearchTrackResponse{}, err
	}

	defer resp.Body.Close()

	collection, err := decode.JSON[models.SearchTrackResponse](resp)
	if err != nil {
		return models.SearchTrackResponse{}, err
	}

	return collection, nil
}

func (c *Client) CurrentUser(ctx context.Context) (models.User, error) {
	resp, err := c.Get(ctx, "/me")
	if err != nil {
		return models.User{}, err
	}

	defer resp.Body.Close()

	user, err := decode.JSON[models.User](resp)
	if err != nil {
		return models.User{}, err
	}

	return user, nil
}

func (c *Client) CreatePlaylist(ctx context.Context, userID string, request models.CreatePlaylistRequest) (models.SimplePlaylist, error) {
	resp, err := c.Post(ctx, fmt.Sprintf("/users/%s/playlists", userID), httpclient.WithJSONBody(request))
	if err != nil {
		return models.SimplePlaylist{}, err
	}

	defer resp.Body.Close()

	playlist, err := decode.JSON[models.SimplePlaylist](resp)
	if err != nil {
		return models.SimplePlaylist{}, err
	}

	return playlist, nil
}

func (c *Client) GetPlaylistTracks(ctx context.Context, playlistID string, limit, offset int) (models.PlaylistTrackPage, error) {
	resp, err := c.Get(ctx, fmt.Sprintf("/playlists/%s/tracks", playlistID), httpclient.WithQuery(map[string]string{
		"limit":  strconv.Itoa(limit),
		"offset": strconv.Itoa(offset),
	}))
	if err != nil {
		return models.PlaylistTrackPage{}, err
	}

	defer resp.Body.Close()

	page, err := decode.JSON[models.PlaylistTrackPage](resp)
	if err != nil {
		return models.PlaylistTrackPage{}, err
	}

	return page, nil
}

func (c *Client) AddItemsToPlaylist(ctx context.Context, playlistID string, request models.AddItemsToPlaylistRequest) (string, error) {
	resp, err := c.Post(ctx, fmt.Sprintf("/playlists/%s/tracks", playlistID), httpclient.WithJSONBody(request))
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	playlist, err := decode.JSON[models.PlaylistSnapshot](resp)
	if err != nil {
		return "", err
	}

	return playlist.SnapshotID, nil
}
