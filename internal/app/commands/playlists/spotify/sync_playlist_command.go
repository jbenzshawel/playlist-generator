package spotify

import (
	"context"
	"log/slog"
	"time"

	"github.com/jbenzshawel/playlist-generator/internal/app/commands/playlists/spotify/internal/providers"
	"github.com/jbenzshawel/playlist-generator/internal/app/commands/playlists/spotify/models"
	"github.com/jbenzshawel/playlist-generator/internal/common/dateformat"
	"github.com/jbenzshawel/playlist-generator/internal/common/decorator"
	"github.com/jbenzshawel/playlist-generator/internal/domain"
)

type SyncPlaylistCommand struct {
	Playlist domain.Playlist
	Date     string
}

type SyncPlaylistCommandHandler decorator.CommandHandler[SyncPlaylistCommand]

func NewSyncPlaylistCommand(
	playlist playlist,
	repository domain.Repository,
) SyncPlaylistCommandHandler {
	return decorator.ApplyDBTransactionDecorator(
		&syncPlaylistCommandHandler{
			provider:           providers.NewPlaylistTrackProvider(playlist),
			playlist:           playlist,
			playlistRepository: repository.Playlist(),
			trackRepository:    repository.SpotifyTrack(),
		},
		repository,
	)
}

type playlistTrackProvider interface {
	GetTracks(ctx context.Context, playlistID string) ([]models.SimpleTrack, error)
}

type playlist interface {
	GetPlaylistTracks(ctx context.Context, playlistID string, limit, offset int) (models.PlaylistTrackPage, error)
	AddItemsToPlaylist(ctx context.Context, playlistID string, request models.AddItemsToPlaylistRequest) (string, error)
}

type syncPlaylistCommandHandler struct {
	provider           playlistTrackProvider
	playlist           playlist
	playlistRepository domain.PlaylistRepository
	trackRepository    domain.SpotifyTrackRepository
}

func (c *syncPlaylistCommandHandler) Execute(ctx context.Context, cmd SyncPlaylistCommand) (any, error) {
	startDate := cmd.Playlist.LastDaySynced()
	if cmd.Date < cmd.Playlist.LastDaySynced() {
		startDate = cmd.Playlist.StartDate()
	}

	endDate, err := cmd.Playlist.EndDate()
	if err != nil {
		return nil, err
	}

	tracks, err := c.trackRepository.GetTracksPlayedInRange(ctx, domain.StudioOneSourceType, startDate, endDate)
	if err != nil {
		return nil, err
	}

	if len(tracks) == 0 {
		slog.Info("no new downloaded tracks to sync")
	}

	trackURIs, err := c.getTrackURIs(ctx, cmd.Playlist, tracks)
	if err != nil {
		return nil, err
	}

	if len(trackURIs) == 0 {
		slog.Info("all downloaded tracks synced to playlist")
	}

	// spotify API supports adding songs with a max batch size of 100
	limit := 100
	for offset := 0; offset < len(trackURIs); offset += limit {
		end := offset + limit

		if end > len(trackURIs) {
			end = len(trackURIs)
		}

		batch := trackURIs[offset:end]

		_, err = c.playlist.AddItemsToPlaylist(ctx, cmd.Playlist.ID(), models.AddItemsToPlaylistRequest{
			URIs: batch,
		})
		if err != nil {
			return nil, err
		}
	}

	// Set last date synced to yesterday since we want to pick up other songs from today
	syncDate := time.Now().AddDate(0, 0, -1).Format(dateformat.YearMonthDay)
	err = c.playlistRepository.SetLastDaySynced(ctx, cmd.Playlist.ID(), syncDate)
	if err != nil {
		return nil, err
	}

	slog.Info("tracks sync complete", slog.Int("numTracks", len(trackURIs)))

	return nil, nil
}

func (c *syncPlaylistCommandHandler) getTrackURIs(ctx context.Context, p domain.Playlist, tracks []domain.SpotifyTrack) ([]string, error) {
	playlistTracks, err := c.provider.GetTracks(ctx, p.ID())
	if err != nil {
		return nil, err
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
	return trackURIs, nil
}
