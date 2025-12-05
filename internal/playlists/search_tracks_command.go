package playlists

import (
	"context"
	"fmt"
	"log/slog"
	"sync/atomic"

	"golang.org/x/sync/errgroup"

	"github.com/jbenzshawel/playlist-generator/internal/common/decorator"
	"github.com/jbenzshawel/playlist-generator/internal/common/output"
	"github.com/jbenzshawel/playlist-generator/internal/domain"
	"github.com/jbenzshawel/playlist-generator/internal/playlists/services"
)

type SearchTracksCommand struct {
	Progress output.ProgressBarCreator
}

type SearchTracksCommandResult struct {
	UnknownCount int
	MatchedCount int
}

type SearchTracksCommandHandler decorator.CommandWithResultHandler[SearchTracksCommand, SearchTracksCommandResult]

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

func (t *searchTracksCommandHandler) Execute(ctx context.Context, cmd SearchTracksCommand) (SearchTracksCommandResult, error) {
	songs, err := t.repository.GetUnknownSongs(ctx)
	if err != nil {
		return SearchTracksCommandResult{}, err
	}

	slog.Info("found unknown songs to search", slog.Int("numSongs", len(songs)))

	if len(songs) == 0 {
		return SearchTracksCommandResult{UnknownCount: 0, MatchedCount: 0}, nil
	}

	g, gCtx := errgroup.WithContext(ctx)

	g.SetLimit(6)
	total := len(songs)

	tracker := cmd.Progress("Searching songs on Spotify", total)
	defer tracker.Stop()

	matchCount := atomic.Int32{}

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

				if track.MatchFound() {
					matchCount.Add(1)
				}

				tracker.Increment()

				return nil
			}
		})
	}

	err = g.Wait()
	if err != nil {
		return SearchTracksCommandResult{}, err
	}

	return SearchTracksCommandResult{UnknownCount: total, MatchedCount: int(matchCount.Load())}, nil
}
