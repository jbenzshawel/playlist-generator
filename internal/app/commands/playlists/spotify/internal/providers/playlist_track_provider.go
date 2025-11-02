package providers

import (
	"context"
	"log/slog"

	"golang.org/x/sync/errgroup"

	"github.com/jbenzshawel/playlist-generator/internal/app/commands/playlists/spotify/models"
)

const maxPageSize = 50

type PlaylistTrackGetter interface {
	GetPlaylistTracks(ctx context.Context, playlistID string, limit, offset int) (models.PlaylistTrackPage, error)
}

func NewPlaylistTrackProvider(getter PlaylistTrackGetter) *playlistTrackProvider {
	return &playlistTrackProvider{
		getter: getter,
	}
}

type playlistTrackProvider struct {
	getter PlaylistTrackGetter
}

func (p *playlistTrackProvider) GetTracks(ctx context.Context, playlistID string) ([]models.SimpleTrack, error) {
	page, err := p.getter.GetPlaylistTracks(ctx, playlistID, maxPageSize, 0)
	if err != nil {
		return nil, err
	}

	slog.Debug("retrieving tracks for playlist", slog.Int("total", page.Total))

	var tracks []models.SimpleTrack

	tracks = append(tracks, pageTracks(page)...)

	if page.Total <= maxPageSize {
		return tracks, nil
	}

	g, gCtx := errgroup.WithContext(ctx)
	g.SetLimit(6) // Set limit to prevent being rate limited

	pages := (page.Total + maxPageSize - 1) / maxPageSize

	pageResults := make(chan []models.SimpleTrack, pages)

	// start idx at 1 since we've already loaded the first page
	for idx := 1; idx < pages; idx++ {
		g.Go(func() error {
			offset := idx * maxPageSize
			reqPage, err := p.getter.GetPlaylistTracks(gCtx, playlistID, maxPageSize, offset)
			if err != nil {
				return err
			}

			select {
			case <-gCtx.Done():
				return gCtx.Err()
			case pageResults <- pageTracks(reqPage):
				return nil
			}
		})
	}

	go func() {
		g.Wait() // Wait for all workers in the group to finish
		close(pageResults)
	}()

	for pr := range pageResults {
		tracks = append(tracks, pr...)
	}

	// Return any error from the worker pool
	if err := g.Wait(); err != nil {
		return nil, err
	}

	return tracks, nil
}

func pageTracks(page models.PlaylistTrackPage) []models.SimpleTrack {
	tracks := make([]models.SimpleTrack, len(page.Items))
	for idx, item := range page.Items {
		tracks[idx] = item.Track
	}

	return tracks
}
