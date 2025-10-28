package spotify

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/jbenzshawel/playlist-generator/internal/app/commands/playlists/spotify/internal/providers"
	"github.com/jbenzshawel/playlist-generator/internal/common/async"
	"github.com/jbenzshawel/playlist-generator/internal/common/decorator"
	"github.com/jbenzshawel/playlist-generator/internal/domain"
)

type SearchTracksCommand struct{}

type SearchTracksCommandHandler decorator.CommandHandler[SearchTracksCommand]

func NewSearchTracksCommand(searcher providers.TrackSearcher, repository domain.Repository) SearchTracksCommandHandler {
	return decorator.ApplyDBTransactionDecorator(
		&searchTracksCommandHandler{
			provider:   providers.NewSearchTrackProvider(searcher),
			repository: repository.SpotifyTrack(),
		},
		repository,
	)
}

type searchProvider interface {
	SearchTrack(ctx context.Context, song domain.Song) (domain.SpotifyTrack, error)
}

type searchTracksCommandHandler struct {
	provider   searchProvider
	repository domain.SpotifyTrackRepository
}

func (t *searchTracksCommandHandler) Execute(ctx context.Context, _ SearchTracksCommand) (any, error) {
	songs, err := t.repository.GetUnknownSongs(ctx)
	if err != nil {
		return nil, err
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

		track, err = t.provider.SearchTrack(ctx, song)
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
		return nil, err
	}

	return nil, nil
}
