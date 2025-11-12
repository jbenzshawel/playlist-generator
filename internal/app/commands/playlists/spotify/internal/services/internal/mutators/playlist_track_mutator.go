package mutators

import (
	"context"

	"github.com/jbenzshawel/playlist-generator/internal/app/commands/playlists/spotify/models"
)

// spotify API supports adding/removing tracks with a max batch size of 100
const batchSize = 100

type TrackAdderRemover interface {
	AddItemsToPlaylist(ctx context.Context, playlistID string, request models.AddItemsToPlaylistRequest) (string, error)
	RemoveItemsFromPlaylist(ctx context.Context, playlistID string, request models.RemoveItemsFromPlaylistRequest) (string, error)
}

type PlaylistTrackMutator interface {
	AddTracks(ctx context.Context, playlistID string, trackURIs []string) error
	RemoveTracks(ctx context.Context, playlistID string, trackURIs []string) error
}

type playlistTrackMutator struct {
	mutator TrackAdderRemover
}

func NewPlaylistTrackMutator(mutator TrackAdderRemover) PlaylistTrackMutator {
	return &playlistTrackMutator{
		mutator: mutator,
	}
}

func (p *playlistTrackMutator) AddTracks(ctx context.Context, playlistID string, trackURIs []string) error {
	err := batchTrackAction(trackURIs, func(batch []string) error {
		_, err := p.mutator.AddItemsToPlaylist(ctx, playlistID, models.AddItemsToPlaylistRequest{
			URIs: batch,
		})
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

func (p *playlistTrackMutator) RemoveTracks(ctx context.Context, playlistID string, trackURIs []string) error {
	err := batchTrackAction(trackURIs, func(batch []string) error {
		tracks := make([]models.RemoveTrack, len(batch))
		for idx, trackURI := range batch {
			tracks[idx] = models.RemoveTrack{
				URI: trackURI,
			}
		}

		_, err := p.mutator.RemoveItemsFromPlaylist(ctx, playlistID, models.RemoveItemsFromPlaylistRequest{
			Tracks: tracks,
		})
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

func batchTrackAction(trackURIs []string, action func(batch []string) error) error {
	var err error
	for offset := 0; offset < len(trackURIs); offset += batchSize {
		end := offset + batchSize
		if end > len(trackURIs) {
			end = len(trackURIs)
		}

		batch := trackURIs[offset:end]

		err = action(batch)
		if err != nil {
			return err
		}
	}

	return nil
}
