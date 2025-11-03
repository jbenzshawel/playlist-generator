package providers

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/jbenzshawel/playlist-generator/internal/app/commands/playlists/spotify/models"
)

func TestPlaylistTrackProvider_GetTracks(t *testing.T) {
	const (
		total      = 201
		playlistID = "testPlaylistID"
	)

	ctx := context.Background()

	testTrackPages := getTestTrackPages(total)

	mockGetter := NewMockPlaylistTrackGetter(t)
	for idx, page := range testTrackPages {
		mockGetter.EXPECT().GetPlaylistTracks(mock.Anything, playlistID, maxPageSize, idx*maxPageSize).Return(page, nil)
	}

	p := NewPlaylistTrackProvider(mockGetter)

	actualTracks, err := p.GetTracks(ctx, playlistID)
	require.NoError(t, err)

	trackIDs := map[string]struct{}{}
	for _, track := range actualTracks {
		trackIDs[track.ID] = struct{}{}
	}
	assert.Len(t, trackIDs, total)
}

func getTestTrackPages(numTracks int) []models.PlaylistTrackPage {
	numPages := (numTracks + maxPageSize - 1) / maxPageSize

	playlistPages := make([]models.PlaylistTrackPage, numPages)
	for pageIdx := 0; pageIdx < numPages; pageIdx++ {
		pageSize := maxPageSize
		remaining := numTracks - (pageIdx+1)*maxPageSize
		if remaining < 0 {
			pageSize = maxPageSize + remaining
		}

		for trackIdx := 0; trackIdx < pageSize; trackIdx++ {
			playlistPages[pageIdx].Total = numTracks
			playlistPages[pageIdx].Items = append(playlistPages[pageIdx].Items, models.PlaylistItem{
				Track: models.SimpleTrack{
					ID: fmt.Sprintf("%d-%d", pageIdx, trackIdx),
				},
			})
		}
	}

	return playlistPages
}
