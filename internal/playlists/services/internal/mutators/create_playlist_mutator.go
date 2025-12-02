package mutators

import (
	"context"
	"time"

	"github.com/jbenzshawel/playlist-generator/internal/common/dateformat"
	"github.com/jbenzshawel/playlist-generator/internal/domain"
	"github.com/jbenzshawel/playlist-generator/internal/playlists/models"
)

type PlaylistCreator interface {
	CurrentUser(ctx context.Context) (models.User, error)
	CreatePlaylist(ctx context.Context, userID string, request models.CreatePlaylistRequest) (models.SimplePlaylist, error)
}

type CreatePlaylistMutator interface {
	CreatePlaylist(ctx context.Context, name string, sourceType domain.SourceType, date time.Time) (domain.Playlist, error)
}

func NewCreatePlaylistMutator(creator PlaylistCreator) CreatePlaylistMutator {
	return &createPlaylistMutator{
		creator: creator,
	}
}

type createPlaylistMutator struct {
	creator PlaylistCreator
}

func (c *createPlaylistMutator) CreatePlaylist(ctx context.Context, name string, sourceType domain.SourceType, date time.Time) (domain.Playlist, error) {
	u, err := c.creator.CurrentUser(ctx)
	if err != nil {
		return domain.Playlist{}, err
	}

	playlistDate := date.Format(dateformat.YearMonth)

	spotifyPlaylist, err := c.creator.CreatePlaylist(ctx, u.ID, models.CreatePlaylistRequest{
		Name: name,
	})
	if err != nil {
		return domain.Playlist{}, err
	}

	p := domain.NewPlaylist(
		spotifyPlaylist.ID,
		spotifyPlaylist.URI,
		spotifyPlaylist.Name,
		playlistDate,
		domain.SpotifyPlaylistType,
		sourceType,
	)

	return p, nil
}
