package providers

import (
	"context"
	"log/slog"

	"golang.org/x/sync/errgroup"

	"github.com/jbenzshawel/playlist-generator/internal/app/commands/playlists/spotify/models"
)

const maxPageSize = 50

type TrackGetter interface {
	GetPlaylistTracks(ctx context.Context, playlistID string, limit, offset int) (models.PlaylistTrackPage, error)
}

type PlaylistTrackProvider interface {
	GetTracks(ctx context.Context, playlistID string) ([]models.SimpleTrack, error)
}

func NewPlaylistTrackProvider(getter TrackGetter) PlaylistTrackProvider {
	return &playlistTrackProvider{
		getter: getter,
	}
}

type playlistTrackProvider struct {
	getter TrackGetter
}

func (p *playlistTrackProvider) GetTracks(ctx context.Context, playlistID string) ([]models.SimpleTrack, error) {
	page, err := p.getter.GetPlaylistTracks(ctx, playlistID, maxPageSize, 0)
	if err != nil {
		return nil, err
	}

	if page.Total == 0 {
		return []models.SimpleTrack{}, nil
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

	pageResults := make([][]models.SimpleTrack, pages)

	// start idx at 1 since we've already loaded the first page
	for idx := 1; idx < pages; idx++ {
		g.Go(func() error {
			offset := idx * maxPageSize
			reqPage, err := p.getter.GetPlaylistTracks(gCtx, playlistID, maxPageSize, offset)
			if err != nil {
				return err
			}

			pageResults[idx] = pageTracks(reqPage)

			return nil
		})
	}

	// Return any error from the worker pool
	if err := g.Wait(); err != nil {
		return nil, err
	}

	for _, pageResult := range pageResults {
		tracks = append(tracks, pageResult...)
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
