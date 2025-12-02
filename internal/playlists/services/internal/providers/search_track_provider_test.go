package providers

import (
	"context"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2/clientcredentials"

	"github.com/jbenzshawel/playlist-generator/internal/app/config"
	"github.com/jbenzshawel/playlist-generator/internal/clients/spotifyclient"
	"github.com/jbenzshawel/playlist-generator/internal/domain"
	"github.com/jbenzshawel/playlist-generator/internal/playlists/models"
)

func TestSearchTrackProvider_SearchTrack(t *testing.T) {
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
		name          string
		song          domain.Song
		searchResults models.SearchTrackResponse
		expectedTrack domain.SpotifyTrack
		expectedErr   error
	}{
		{
			name: "single result",
			song: song,
			searchResults: models.SearchTrackResponse{
				Tracks: models.TrackCollection{
					Total: 1,
					Items: []models.SimpleTrack{
						{
							Album: models.Album{
								AlbumType: models.AlbumAlbumType,
								Name:      album,
							},
							Artists: []models.Artist{
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
			searchResults: models.SearchTrackResponse{
				Tracks: models.TrackCollection{
					Total: 3,
					Items: []models.SimpleTrack{
						{
							Album: models.Album{
								AlbumType: models.AlbumAlbumType,
								Name:      "Never There (single)",
							},
							Artists: []models.Artist{
								{
									Name: artist,
								},
							},
							Name: track,
							ID:   "single",
							URI:  "single",
						},
						{
							Album: models.Album{
								AlbumType: models.AlbumAlbumType,
								Name:      "Never There (single)",
							},
							Artists: []models.Artist{
								{
									Name: artist,
								},
							},
							Name: "Never There (single)",
							ID:   "single",
							URI:  "single",
						},
						{
							Album: models.Album{
								AlbumType: models.AlbumAlbumType,
								Name:      "Prolonging Magic",
							},
							Artists: []models.Artist{
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
			searchResults: models.SearchTrackResponse{
				Tracks: models.TrackCollection{
					Total: 2,
					Items: []models.SimpleTrack{
						{
							Album: models.Album{
								AlbumType: models.AlbumAlbumType,
								Name:      album,
							},
							Artists: []models.Artist{
								{
									Name: artist,
								},
							},
							Name: track,
							ID:   "no match",
							URI:  "no match",
						},
						{
							Album: models.Album{
								AlbumType: models.SingleAlbumType,
								Name:      "single",
							},
							Artists: []models.Artist{
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
			name: "match at min threshold ",
			song: song,
			searchResults: models.SearchTrackResponse{
				Tracks: models.TrackCollection{
					Total: 2,
					Items: []models.SimpleTrack{
						{
							Album: models.Album{
								AlbumType: models.AlbumAlbumType,
								Name:      "no match",
							},
							Artists: []models.Artist{
								{
									Name: "no match",
								},
							},
							Name: "no match",
							ID:   "no match",
							URI:  "no match",
						},
						{
							Album: models.Album{
								AlbumType: models.AlbumAlbumType,
								Name:      "mag",
							},
							Artists: []models.Artist{
								{
									Name: "cake",
								},
							},
							Name: "Never Ther",
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
			searchResults: models.SearchTrackResponse{
				Tracks: models.TrackCollection{
					Total: 2,
					Items: []models.SimpleTrack{
						{
							Album: models.Album{
								AlbumType: models.AlbumAlbumType,
								Name:      "no match",
							},
							Artists: []models.Artist{
								{
									Name: "no match",
								},
							},
							Name: "no match",
							ID:   "no match",
							URI:  "no match",
						},
						{
							Album: models.Album{
								AlbumType: models.AlbumAlbumType,
								Name:      "ma",
							},
							Artists: []models.Artist{
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
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()

			searcher := NewMockTrackSearcher(t)
			searcher.EXPECT().SearchTrack(ctx, tc.song.Artist(), tc.song.Track(), "").Return(tc.searchResults, nil)

			provider := searchTrackProvider{
				searcher: searcher,
			}

			actualTrack, err := provider.SearchTrack(ctx, tc.song)
			assert.ErrorIs(t, err, tc.expectedErr)
			assert.Equal(t, tc.expectedTrack, actualTrack)
		})
	}
}

func TestSearchTrackProvider_SearchTrack_Integration(t *testing.T) {
	// This test actualy calls the sptofiy API and is used to troubleshoot cases
	// where a match isn't found
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	const (
		artist = "Declan Mckenna"
		track  = "Twice Your Size"
		album  = "Zeros"
	)

	appConfig := config.LoadForTest(t)

	oauthConfig := clientcredentials.Config{
		ClientID:     appConfig.SpotifyClient.ClientID,
		ClientSecret: appConfig.SpotifyClient.ClientSecret,
		TokenURL:     appConfig.SpotifyClient.TokenURL,
	}

	baseURL, err := url.Parse(appConfig.SpotifyClient.BaseURL)
	require.NoError(t, err)

	spotifyClient := spotifyclient.New(spotifyclient.Config{
		BaseURL: baseURL,
		Client:  oauthConfig.Client(t.Context()),
	})

	provider := searchTrackProvider{
		searcher: spotifyClient,
	}

	s, err := domain.NewSong(artist, track, album, "")
	require.NoError(t, err)

	_, err = provider.SearchTrack(t.Context(), s)
	require.NoError(t, err)
}
