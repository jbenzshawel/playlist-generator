package providers

import (
	"context"
	"github.com/jbenzshawel/playlist-generator/internal/app/playlists/spotify"
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
		searchResults      spotify.SearchTrackResponse
		searchWithoutAlbum bool
		expectedTrack      domain.SpotifyTrack
		expectedErr        error
	}{
		{
			name: "single result",
			song: song,
			searchResults: spotify.SearchTrackResponse{
				Tracks: spotify.TrackCollection{
					Total: 1,
					Items: []spotify.Track{
						{
							Album: spotify.Album{
								AlbumType: spotify.AlbumAlbumType,
								Name:      album,
							},
							Artists: []spotify.Artist{
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
			searchResults: spotify.SearchTrackResponse{
				Tracks: spotify.TrackCollection{
					Total: 3,
					Items: []spotify.Track{
						{
							Album: spotify.Album{
								AlbumType: spotify.AlbumAlbumType,
								Name:      "Never There (single)",
							},
							Artists: []spotify.Artist{
								{
									Name: artist,
								},
							},
							Name: track,
							ID:   "single",
							URI:  "single",
						},
						{
							Album: spotify.Album{
								AlbumType: spotify.AlbumAlbumType,
								Name:      "Never There (single)",
							},
							Artists: []spotify.Artist{
								{
									Name: artist,
								},
							},
							Name: "Never There (single)",
							ID:   "single",
							URI:  "single",
						},
						{
							Album: spotify.Album{
								AlbumType: spotify.AlbumAlbumType,
								Name:      "Prolonging Magic",
							},
							Artists: []spotify.Artist{
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
			searchResults: spotify.SearchTrackResponse{
				Tracks: spotify.TrackCollection{
					Total: 2,
					Items: []spotify.Track{
						{
							Album: spotify.Album{
								AlbumType: spotify.AlbumAlbumType,
								Name:      album,
							},
							Artists: []spotify.Artist{
								{
									Name: artist,
								},
							},
							Name: track,
							ID:   "no match",
							URI:  "no match",
						},
						{
							Album: spotify.Album{
								AlbumType: spotify.SingleAlbumType,
								Name:      "single",
							},
							Artists: []spotify.Artist{
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
			searchResults: spotify.SearchTrackResponse{
				Tracks: spotify.TrackCollection{
					Total: 1,
					Items: []spotify.Track{
						{
							Album: spotify.Album{
								AlbumType: spotify.AlbumAlbumType,
								Name:      album,
							},
							Artists: []spotify.Artist{
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
			searchResults: spotify.SearchTrackResponse{
				Tracks: spotify.TrackCollection{
					Total: 2,
					Items: []spotify.Track{
						{
							Album: spotify.Album{
								AlbumType: spotify.AlbumAlbumType,
								Name:      "no match",
							},
							Artists: []spotify.Artist{
								{
									Name: "no match",
								},
							},
							Name: "no match",
							ID:   "no match",
							URI:  "no match",
						},
						{
							Album: spotify.Album{
								AlbumType: spotify.AlbumAlbumType,
								Name:      "mag",
							},
							Artists: []spotify.Artist{
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
			searchResults: spotify.SearchTrackResponse{
				Tracks: spotify.TrackCollection{
					Total: 2,
					Items: []spotify.Track{
						{
							Album: spotify.Album{
								AlbumType: spotify.AlbumAlbumType,
								Name:      "no match",
							},
							Artists: []spotify.Artist{
								{
									Name: "no match",
								},
							},
							Name: "no match",
							ID:   "no match",
							URI:  "no match",
						},
						{
							Album: spotify.Album{
								AlbumType: spotify.AlbumAlbumType,
								Name:      "ma",
							},
							Artists: []spotify.Artist{
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
			searchResults: spotify.SearchTrackResponse{
				Tracks: spotify.TrackCollection{
					Total: 0,
					Items: []spotify.Track{},
				},
			},
			expectedErr: errTrackNotFound,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()

			searcher := spotify.newMocktrackSearcher(t)
			if tc.searchWithoutAlbum {
				searcher.EXPECT().SearchTrack(ctx, tc.song.Artist(), tc.song.Track(), tc.song.Album()).Return(spotify.SearchTrackResponse{}, nil)
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
