package spotify

import (
	"context"
	"log/slog"

	"github.com/jbenzshawel/playlist-generator/internal/common/async"
	"github.com/jbenzshawel/playlist-generator/internal/domain"
)

type provider interface {
	GetTrack(ctx context.Context, song domain.Song) (domain.SpotifyTrack, error)
}

func NewTrackUpdater(searcher TrackSearcher, r domain.SpotifyTrackRepository) *trackUpdater {
	return &trackUpdater{
		provider:   NewSpotifyTrackProvider(searcher),
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

	// anything higher than 6 workers starts to get rate limited
	async.ParallelFor(ctx, len(songs), 6, func(idx int) {
		defer func() {
			if r := recover(); r != nil {
				slog.Error("recovered from panic searching for tracks", slog.Any("panic", r))
			}
		}()

		var track domain.SpotifyTrack
		var err error
		song := songs[idx]

		track, err = t.provider.GetTrack(ctx, song)
		if err != nil {
			slog.Warn("spotify track not found for song",
				slog.Any("song", song),
				slog.Any("error", err),
			)
			track = domain.NewNotFoundSpotifyTrack(song.ID())
		}

		err = t.repository.Insert(ctx, track)
		if err != nil {
			slog.Warn("error inserting spotify track", slog.Any("error", err))
		}
	})

	return nil
}
