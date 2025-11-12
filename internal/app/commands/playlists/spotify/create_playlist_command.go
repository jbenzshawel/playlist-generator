package spotify

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/jbenzshawel/playlist-generator/internal/app/commands/playlists/spotify/internal/services"
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
	playlistService services.PlaylistService,
	repository domain.Repository,
) CreatePlaylistCommandHandler {
	return decorator.ApplyDBTransactionDecorator(
		&createPlaylistCommand{
			playlistService:    playlistService,
			playlistRepository: repository.Playlist(),
		},
		repository,
	)
}

type createPlaylistCommand struct {
	playlistService    services.PlaylistService
	playlistRepository domain.PlaylistRepository
}

func (c *createPlaylistCommand) Execute(ctx context.Context, cmd CreatePlaylistCommand) (CreatePlaylistCommandResult, error) {
	date, err := time.Parse(time.DateOnly, cmd.Date)
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

	p, err = c.playlistService.CreatePlaylist(ctx, fmt.Sprintf("Studio One %s", playlistDate), date)
	if err != nil {
		return CreatePlaylistCommandResult{}, err
	}

	err = c.playlistRepository.Insert(ctx, p)
	if err != nil {
		return CreatePlaylistCommandResult{}, err
	}

	slog.Info("new spotify playlist created", slog.Any("playlist", p))

	return CreatePlaylistCommandResult{Playlist: p}, nil
}
