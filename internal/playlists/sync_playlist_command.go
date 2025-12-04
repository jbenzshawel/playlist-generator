package playlists

import (
	"context"
	"log/slog"
	"time"

	"github.com/jbenzshawel/playlist-generator/internal/common/decorator"
	"github.com/jbenzshawel/playlist-generator/internal/domain"
	"github.com/jbenzshawel/playlist-generator/internal/playlists/models"
	"github.com/jbenzshawel/playlist-generator/internal/playlists/services"
)

type SyncPlaylistCommand struct {
	Playlist   domain.Playlist
	SourceType domain.SourceType
	Date       string
}

type SyncPlaylistCommandResult struct {
	TotalTracks int
	NewTracks   int
}

type SyncPlaylistCommandHandler decorator.CommandWithResultHandler[SyncPlaylistCommand, SyncPlaylistCommandResult]

func NewSyncPlaylistCommand(
	playlistService services.PlaylistService,
	repository domain.Repository,
) SyncPlaylistCommandHandler {
	return decorator.ApplyDBTransactionDecorator(
		&syncPlaylistCommandHandler{
			playlistService:    playlistService,
			playlistRepository: repository.Playlist(),
			trackRepository:    repository.SpotifyTrack(),
		},
		repository,
	)
}

type playlist interface {
	AddItemsToPlaylist(ctx context.Context, playlistID string, request models.AddItemsToPlaylistRequest) (string, error)
}

type syncPlaylistCommandHandler struct {
	playlistService    services.PlaylistService
	playlistRepository domain.PlaylistRepository
	trackRepository    domain.SpotifyTrackRepository
}

func (c *syncPlaylistCommandHandler) Execute(ctx context.Context, cmd SyncPlaylistCommand) (SyncPlaylistCommandResult, error) {
	startDate := cmd.Playlist.LastDaySynced()
	if startDate == "" || cmd.Date < cmd.Playlist.LastDaySynced() {
		startDate = cmd.Playlist.StartDate()
	}

	endDate, err := cmd.Playlist.EndDate()
	if err != nil {
		return SyncPlaylistCommandResult{}, err
	}

	tracks, err := c.trackRepository.GetTracksPlayedInRange(ctx, cmd.SourceType, startDate, endDate)
	if err != nil {
		return SyncPlaylistCommandResult{}, err
	}

	if len(tracks) == 0 {
		slog.Info("no new downloaded tracks to sync")
	}

	trackURIs, playlistTotal, err := c.getTrackURIs(ctx, cmd.Playlist, tracks)
	if err != nil {
		return SyncPlaylistCommandResult{}, err
	}

	if len(trackURIs) == 0 {
		slog.Info("all downloaded tracks synced to playlist")
	}

	err = c.playlistService.AddTracks(ctx, cmd.Playlist.ID(), trackURIs)
	if err != nil {
		return SyncPlaylistCommandResult{}, err
	}

	// Set last date synced to yesterday since we want to pick up other songs from today
	syncDate := time.Now().AddDate(0, 0, -1).Format(time.DateOnly)
	err = c.playlistRepository.SetLastDaySynced(ctx, cmd.Playlist.ID(), syncDate)
	if err != nil {
		return SyncPlaylistCommandResult{}, err
	}

	slog.Info("tracks sync complete", slog.Int("numTracks", len(trackURIs)))

	return SyncPlaylistCommandResult{TotalTracks: playlistTotal, NewTracks: len(trackURIs)}, nil
}

func (c *syncPlaylistCommandHandler) getTrackURIs(ctx context.Context, p domain.Playlist, tracks []domain.SpotifyTrack) ([]string, int, error) {
	playlistTracks, err := c.playlistService.GetTracks(ctx, p.ID())
	if err != nil {
		return nil, 0, err
	}

	trackLookup := make(map[string]struct{}, len(playlistTracks))
	for _, track := range playlistTracks {
		trackLookup[track.ID] = struct{}{}
	}

	trackURIs := make([]string, 0, len(tracks))
	for _, track := range tracks {
		if _, ok := trackLookup[track.TrackID()]; !ok {
			trackURIs = append(trackURIs, track.URI())
		}
	}
	return trackURIs, len(playlistTracks), nil
}
