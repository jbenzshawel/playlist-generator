package providers

import (
	"context"
	"golang.org/x/sync/errgroup"
	"sync"

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

	var tracks []models.SimpleTrack

	tracks = append(tracks, pageTracks(page)...)

	if page.Total < maxPageSize {
		return tracks, nil
	}

	var l sync.Mutex
	g, gCtx := errgroup.WithContext(ctx)
	g.SetLimit(6)

	pages := page.Total / maxPageSize
	// start idx at 1 since we've already loaded the first page
	for idx := 1; idx < pages; idx++ {
		g.Go(func() error {
			l.Lock()
			defer l.Unlock()

			offset := idx * maxPageSize
			reqPage, err := p.getter.GetPlaylistTracks(gCtx, playlistID, maxPageSize, offset)
			if err != nil {
				return err
			}

			tracks = append(tracks, pageTracks(reqPage)...)

			return nil
		})
	}

	err = g.Wait()
	if err != nil {
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
