package spotify

import (
	"context"
	"fmt"
	"github.com/jbenzshawel/playlist-generator/internal/common/decorator"
	"github.com/jbenzshawel/playlist-generator/internal/domain"
	"time"
)

const dayFormat = "2006-01-02"

type CreatePlaylistCommand struct {
	Date string
}

type CreatePlaylistCommandResult struct {
	Playlist domain.Playlist
}

type CreatePlaylistCommandHandler decorator.CommandWithResultHandler[CreatePlaylistCommand, CreatePlaylistCommandResult]

func NewCreatePlaylistCommand(
	creator creator,
	repository domain.Repository,
) CreatePlaylistCommandHandler {
	return decorator.ApplyDBTransactionDecorator(
		&createPlaylistCommand{
			creator:            creator,
			playlistRepository: repository.Playlist(),
		},
		repository,
	)
}

type creator interface {
	CurrentUser(ctx context.Context) (User, error)
	CreatePlaylist(ctx context.Context, userID string, request CreatePlaylistRequest) (SimplePlaylist, error)
}

type createPlaylistCommand struct {
	creator            creator
	playlistRepository domain.PlaylistRepository
}

func (c *createPlaylistCommand) Execute(ctx context.Context, cmd CreatePlaylistCommand) (CreatePlaylistCommandResult, error) {
	date, err := time.Parse(dayFormat, cmd.Date)
	if err != nil {
		return CreatePlaylistCommandResult{}, fmt.Errorf("invalid create playlist date: %w", err)
	}

	p, err := c.playlistRepository.GetPlaylistByMonth(ctx, domain.SpotifyPlaylistType, date.Year(), date.Month())
	if err != nil {
		return CreatePlaylistCommandResult{}, err
	}

	if !p.IsZero() {
		return CreatePlaylistCommandResult{Playlist: p}, nil
	}

	u, err := c.creator.CurrentUser(ctx)
	if err != nil {
		return CreatePlaylistCommandResult{}, err
	}

	spotifyPlaylist, err := c.creator.CreatePlaylist(ctx, u.ID, CreatePlaylistRequest{
		Name: fmt.Sprintf("Studio One %d-%d", date.Year(), date.Month()),
	})
	if err != nil {
		return CreatePlaylistCommandResult{}, err
	}

	p = domain.NewPlaylist(spotifyPlaylist.ID, spotifyPlaylist.URI, spotifyPlaylist.Name, cmd.Date, domain.SpotifyPlaylistType, domain.StudioOneSourceType)

	err = c.playlistRepository.Insert(ctx, p)
	if err != nil {
		return CreatePlaylistCommandResult{}, err
	}

	return CreatePlaylistCommandResult{Playlist: p}, nil
}
