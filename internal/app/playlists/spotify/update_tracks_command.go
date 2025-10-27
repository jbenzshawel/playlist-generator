package spotify

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/jbenzshawel/playlist-generator/internal/common/async"
	"github.com/jbenzshawel/playlist-generator/internal/common/decorator"
	"github.com/jbenzshawel/playlist-generator/internal/domain"
)

type UpdateTracksCommand struct{}

type UpdateTracksCommandHandler decorator.CommandHandler[UpdateTracksCommand]

func NewUpdateTracksCommandHandler(searcher TrackSearcher, repository domain.Repository) UpdateTracksCommandHandler {
	return decorator.ApplyDBTransactionDecorator(
		&trackUpdateCommand{
			provider:   NewSpotifyTrackProvider(searcher),
			repository: repository.SpotifyTracks(),
		},
		repository,
	)
}

type provider interface {
	GetTrack(ctx context.Context, song domain.Song) (domain.SpotifyTrack, error)
}

type trackUpdateCommand struct {
	provider   provider
	repository domain.SpotifyTrackRepository
}

func (t *trackUpdateCommand) Execute(ctx context.Context, _ UpdateTracksCommand) error {
	songs, err := t.repository.GetUnknownSongs(ctx)
	if err != nil {
		return err
	}

	// anything higher than 6 workers starts to get rate limited
	err = async.ParallelFor(ctx, len(songs), 6, func(ctx context.Context, idx int) error {
		defer func() {
			if r := recover(); r != nil {
				err = fmt.Errorf("panic occurred during searching for tracks: %v", r)
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
			return fmt.Errorf("spotify track insert error: %w", err)
		}

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}
