package spotify

import (
	"context"
	"fmt"
	"log/slog"

	"golang.org/x/sync/errgroup"

	"github.com/jbenzshawel/playlist-generator/internal/app/commands/playlists/spotify/internal/services"
	"github.com/jbenzshawel/playlist-generator/internal/common/decorator"
	"github.com/jbenzshawel/playlist-generator/internal/domain"
)

type SearchTracksCommand struct{}

type SearchTracksCommandHandler decorator.CommandHandler[SearchTracksCommand]

func NewSearchTracksCommand(searchService services.SearchService, repository domain.Repository) SearchTracksCommandHandler {
	return decorator.ApplyDBTransactionDecorator(
		&searchTracksCommandHandler{
			searchService: searchService,
			repository:    repository.SpotifyTrack(),
		},
		repository,
	)
}

type searchTracksCommandHandler struct {
	searchService services.SearchService
	repository    domain.SpotifyTrackRepository
}

func (t *searchTracksCommandHandler) Execute(ctx context.Context, _ SearchTracksCommand) (any, error) {
	songs, err := t.repository.GetUnknownSongs(ctx)
	if err != nil {
		return nil, err
	}

	slog.Info("found unknown songs to search", slog.Int("numSongs", len(songs)))

	g, gCtx := errgroup.WithContext(ctx)

	g.SetLimit(6)

	for idx := 0; idx < len(songs); idx++ {
		g.Go(func() error {
			defer func() {
				if r := recover(); r != nil {
					err = fmt.Errorf("panic occurred during searching for tracks: %v", r)
				}
			}()

			select {
			case <-gCtx.Done():
				return gCtx.Err()
			default:
				var track domain.SpotifyTrack
				var err error
				song := songs[idx]

				track, err = t.searchService.SearchTrack(ctx, song)
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
			}
		})
	}

	err = g.Wait()
	if err != nil {
		return nil, err
	}

	return nil, nil
}
