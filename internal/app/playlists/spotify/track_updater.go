package spotify

import (
	"context"
	"github.com/jbenzshawel/playlist-generator/internal/app/playlists/spotify/internal/providers"
	"log/slog"

	"github.com/jbenzshawel/playlist-generator/internal/domain"
)

type provider interface {
	GetTrack(ctx context.Context, song domain.Song) (domain.SpotifyTrack, error)
}

func NewTrackUpdater(searcher providers.TrackSearcher, r domain.SpotifyTrackRepository) *trackUpdater {
	return &trackUpdater{
		provider:   providers.NewSpotifyTrackProvider(searcher),
		repository: r,
	}
}

type trackUpdater struct {
	provider   provider
	repository domain.SpotifyTrackRepository
}

func (t *trackUpdater) UpdateSpotifyTracks(ctx context.Context) error {
	songs, err := t.repository.GetUnknownSongs(ctx)
	if err != nil {
		return err
	}

	for _, song := range songs {
		track, err := t.provider.GetTrack(ctx, song)
		if err != nil {
			slog.Warn("spotify track not found for song", slog.Any("song", song))
			continue
		}

		err = t.repository.Insert(ctx, track)
		if err != nil {
			return err
		}
	}

	return nil
}
