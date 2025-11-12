package spotify

import (
	"context"
	"log/slog"

	"github.com/jbenzshawel/playlist-generator/internal/app/commands/playlists/spotify/internal/services"
	"github.com/jbenzshawel/playlist-generator/internal/common/decorator"
	"github.com/jbenzshawel/playlist-generator/internal/domain"
)

const (
	// TODO: Some day generate this playlist? There will only ever be one so it is
	// probably not worth the squeeze
	randomStudioOnePlaylistID = "533ho0RGEouDWd3RfZc3rw"
)

type RandomTracksPlaylistCommand struct {
	NumTracks int
}

type RandomTracksPlaylistCommandHandler decorator.CommandHandler[RandomTracksPlaylistCommand]

func NewRandomTracksPlaylistCommand(
	playlistService services.PlaylistService,
	repository domain.Repository,
) RandomTracksPlaylistCommandHandler {
	return decorator.ApplyDBTransactionDecorator(
		&randomTracksPlaylistCommand{
			playlistService: playlistService,
			repository:      repository.SpotifyTrack(),
		},
		repository,
	)
}

type randomTracksPlaylistCommand struct {
	playlistService services.PlaylistService
	repository      domain.SpotifyTrackRepository
}

func (r *randomTracksPlaylistCommand) Execute(ctx context.Context, cmd RandomTracksPlaylistCommand) (any, error) {
	existingTracks, err := r.playlistService.GetTracks(ctx, randomStudioOnePlaylistID)
	if err != nil {
		return nil, err
	}

	existingTrackURIs := make([]string, len(existingTracks))
	for idx, track := range existingTracks {
		existingTrackURIs[idx] = track.URI
	}

	if len(existingTracks) > 0 {
		err = r.playlistService.RemoveTracks(ctx, randomStudioOnePlaylistID, existingTrackURIs)
		if err != nil {
			return nil, err
		}

		slog.Info("existing tracks removed", slog.Int("numTracks", len(existingTracks)))
	}

	newTracks, err := r.repository.GetRandomTracks(ctx, cmd.NumTracks)
	if err != nil {
		return nil, err
	}

	newTrackURIs := make([]string, len(newTracks))
	for idx, track := range newTracks {
		newTrackURIs[idx] = track.URI()
	}

	err = r.playlistService.AddTracks(ctx, randomStudioOnePlaylistID, newTrackURIs)
	if err != nil {
		return nil, err
	}

	slog.Info("random tracks added", slog.Int("numTracks", len(newTrackURIs)))

	return nil, nil
}
