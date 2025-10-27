package spotify

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jbenzshawel/playlist-generator/internal/domain"
)

func TestSpotifyTrackProvider_GetTrack(t *testing.T) {
	const (
		artist      = "Cake"
		track       = "Never There"
		album       = "Prolonging The Magic"
		albumSingle = "Never There (single)"

		trackID = "7aKWgpecgLEqisWcXPElDl"
		uri     = "spotify:track:7aKWgpecgLEqisWcXPElDl"
	)

	song, err := domain.NewSong(artist, track, album, "")
	require.NoError(t, err)

	songSingle, err := domain.NewSong(artist, track, albumSingle, "")
	require.NoError(t, err)

	testCases := []struct {
		name               string
		song               domain.Song
		searchResults      SearchTrackResponse
		searchWithoutAlbum bool
		expectedTrack      domain.SpotifyTrack
		expectedErr        error
	}{
		{
			name: "single result",
			song: song,
			searchResults: SearchTrackResponse{
				Tracks: TrackCollection{
					Total: 1,
					Items: []Track{
						{
							Album: Album{
								AlbumType: AlbumAlbumType,
								Name:      album,
							},
							Artists: []Artist{
								{
									Name: artist,
								},
							},
							Name: track,
							ID:   trackID,
							URI:  uri,
						},
					},
				},
			},
			expectedTrack: domain.NewSpotifyTrack(song.ID(), trackID, uri),
		},
		{
			name: "multiple partial matches",
			song: song,
			searchResults: SearchTrackResponse{
				Tracks: TrackCollection{
					Total: 3,
					Items: []Track{
						{
							Album: Album{
								AlbumType: AlbumAlbumType,
								Name:      "Never There (single)",
							},
							Artists: []Artist{
								{
									Name: artist,
								},
							},
							Name: track,
							ID:   "single",
							URI:  "single",
						},
						{
							Album: Album{
								AlbumType: AlbumAlbumType,
								Name:      "Never There (single)",
							},
							Artists: []Artist{
								{
									Name: artist,
								},
							},
							Name: "Never There (single)",
							ID:   "single",
							URI:  "single",
						},
						{
							Album: Album{
								AlbumType: AlbumAlbumType,
								Name:      "Prolonging Magic",
							},
							Artists: []Artist{
								{
									Name: "CAKE",
								},
							},
							Name: track,
							ID:   trackID,
							URI:  uri,
						},
					},
				},
			},
			expectedTrack: domain.NewSpotifyTrack(song.ID(), trackID, uri),
		},
		{
			name: "single album type",
			song: songSingle,
			searchResults: SearchTrackResponse{
				Tracks: TrackCollection{
					Total: 2,
					Items: []Track{
						{
							Album: Album{
								AlbumType: AlbumAlbumType,
								Name:      album,
							},
							Artists: []Artist{
								{
									Name: artist,
								},
							},
							Name: track,
							ID:   "no match",
							URI:  "no match",
						},
						{
							Album: Album{
								AlbumType: SingleAlbumType,
								Name:      "single",
							},
							Artists: []Artist{
								{
									Name: artist,
								},
							},
							Name: track,
							ID:   trackID,
							URI:  uri,
						},
					},
				},
			},
			expectedTrack: domain.NewSpotifyTrack(songSingle.ID(), trackID, uri),
		},
		{
			name:               "search without album ",
			song:               song,
			searchWithoutAlbum: true,
			searchResults: SearchTrackResponse{
				Tracks: TrackCollection{
					Total: 1,
					Items: []Track{
						{
							Album: Album{
								AlbumType: AlbumAlbumType,
								Name:      album,
							},
							Artists: []Artist{
								{
									Name: artist,
								},
							},
							Name: track,
							ID:   trackID,
							URI:  uri,
						},
					},
				},
			},
			expectedTrack: domain.NewSpotifyTrack(song.ID(), trackID, uri),
		},
		{
			name: "match at min threshold ",
			song: song,
			searchResults: SearchTrackResponse{
				Tracks: TrackCollection{
					Total: 2,
					Items: []Track{
						{
							Album: Album{
								AlbumType: AlbumAlbumType,
								Name:      "no match",
							},
							Artists: []Artist{
								{
									Name: "no match",
								},
							},
							Name: "no match",
							ID:   "no match",
							URI:  "no match",
						},
						{
							Album: Album{
								AlbumType: AlbumAlbumType,
								Name:      "mag",
							},
							Artists: []Artist{
								{
									Name: "cak",
								},
							},
							Name: "Never There",
							ID:   trackID,
							URI:  uri,
						},
					},
				},
			},
			expectedTrack: domain.NewSpotifyTrack(song.ID(), trackID, uri),
		},
		{
			name: "match below min threshold ",
			song: song,
			searchResults: SearchTrackResponse{
				Tracks: TrackCollection{
					Total: 2,
					Items: []Track{
						{
							Album: Album{
								AlbumType: AlbumAlbumType,
								Name:      "no match",
							},
							Artists: []Artist{
								{
									Name: "no match",
								},
							},
							Name: "no match",
							ID:   "no match",
							URI:  "no match",
						},
						{
							Album: Album{
								AlbumType: AlbumAlbumType,
								Name:      "ma",
							},
							Artists: []Artist{
								{
									Name: "cak",
								},
							},
							Name: "Never There",
							ID:   trackID,
							URI:  uri,
						},
					},
				},
			},
			expectedErr: errMatchBelowThreshold,
		},
		{
			name:               "no results",
			song:               song,
			searchWithoutAlbum: true,
			searchResults: SearchTrackResponse{
				Tracks: TrackCollection{
					Total: 0,
					Items: []Track{},
				},
			},
			expectedErr: errTrackNotFound,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()

			searcher := newMocktrackSearcher(t)
			if tc.searchWithoutAlbum {
				searcher.EXPECT().SearchTrack(ctx, tc.song.Artist(), tc.song.Track(), tc.song.Album()).Return(SearchTrackResponse{}, nil)
				searcher.EXPECT().SearchTrack(ctx, tc.song.Artist(), tc.song.Track(), "").Return(tc.searchResults, nil)
			} else {
				searcher.EXPECT().SearchTrack(ctx, tc.song.Artist(), tc.song.Track(), tc.song.Album()).Return(tc.searchResults, nil)
			}

			provider := spotifyTrackProvider{
				searcher: searcher,
			}

			actualTrack, err := provider.GetTrack(ctx, tc.song)
			assert.ErrorIs(t, err, tc.expectedErr)
			assert.Equal(t, tc.expectedTrack, actualTrack)
		})
	}
}
