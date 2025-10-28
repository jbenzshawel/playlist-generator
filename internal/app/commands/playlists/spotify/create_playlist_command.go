package spotify

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/jbenzshawel/playlist-generator/internal/app/commands/playlists/spotify/models"
	"github.com/jbenzshawel/playlist-generator/internal/common/dateformat"
	"github.com/jbenzshawel/playlist-generator/internal/common/decorator"
	"github.com/jbenzshawel/playlist-generator/internal/domain"
)

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
	CurrentUser(ctx context.Context) (models.User, error)
	CreatePlaylist(ctx context.Context, userID string, request models.CreatePlaylistRequest) (models.SimplePlaylist, error)
}

type createPlaylistCommand struct {
	creator            creator
	playlistRepository domain.PlaylistRepository
}

func (c *createPlaylistCommand) Execute(ctx context.Context, cmd CreatePlaylistCommand) (CreatePlaylistCommandResult, error) {
	date, err := time.Parse(dateformat.YearMonthDay, cmd.Date)
	if err != nil {
		return CreatePlaylistCommandResult{}, fmt.Errorf("invalid create playlist date: %w", err)
	}

	playlistDate := date.Format(dateformat.YearMonth)

	p, err := c.playlistRepository.GetPlaylistByDate(ctx, domain.SpotifyPlaylistType, playlistDate)
	if err != nil {
		return CreatePlaylistCommandResult{}, err
	}

	if !p.IsZero() {
		slog.Info("existing spotify playlist found", slog.Any("playlist", p))
		return CreatePlaylistCommandResult{Playlist: p}, nil
	}

	u, err := c.creator.CurrentUser(ctx)
	if err != nil {
		return CreatePlaylistCommandResult{}, err
	}

	spotifyPlaylist, err := c.creator.CreatePlaylist(ctx, u.ID, models.CreatePlaylistRequest{
		Name: fmt.Sprintf("Studio One %d-%d", date.Year(), date.Month()),
	})
	if err != nil {
		return CreatePlaylistCommandResult{}, err
	}

	p = domain.NewPlaylist(
		spotifyPlaylist.ID,
		spotifyPlaylist.URI,
		spotifyPlaylist.Name,
		playlistDate,
		domain.SpotifyPlaylistType,
		domain.StudioOneSourceType,
	)

	err = c.playlistRepository.Insert(ctx, p)
	if err != nil {
		return CreatePlaylistCommandResult{}, err
	}

	slog.Info("new spotify playlist created", slog.Any("playlist", p))

	return CreatePlaylistCommandResult{Playlist: p}, nil
}
