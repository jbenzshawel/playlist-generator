package providers

import (
	"context"
	"errors"
	"github.com/jbenzshawel/playlist-generator/internal/app/playlists/spotify"
	"log/slog"
	"slices"
	"strings"

	"github.com/jbenzshawel/playlist-generator/internal/common/compare"
	"github.com/jbenzshawel/playlist-generator/internal/domain"
)

const minMatchPercent = 70.0

var (
	errTrackNotFound       = errors.New("track not found")
	errMatchBelowThreshold = errors.New("match below threshold")
)

type TrackSearcher interface {
	SearchTrack(ctx context.Context, artist, track, album string) (spotify.SearchTrackResponse, error)
}

func NewSpotifyTrackProvider(s TrackSearcher) *spotifyTrackProvider {
	return &spotifyTrackProvider{
		searcher: s,
	}
}

type spotifyTrackProvider struct {
	searcher TrackSearcher
}

func (s *spotifyTrackProvider) GetTrack(ctx context.Context, song domain.Song) (domain.SpotifyTrack, error) {
	resp, err := s.searcher.SearchTrack(ctx, song.Artist(), song.Track(), song.Album())
	if err != nil {
		return domain.SpotifyTrack{}, err
	}

	if resp.Tracks.Total > 0 {
		return findSongTrackMatch(resp.Tracks, song)
	}

	slog.Info("track not found with album query param; searching without album", slog.Any("song", song))

	resp, err = s.searcher.SearchTrack(ctx, song.Artist(), song.Track(), "")
	if err != nil {
		return domain.SpotifyTrack{}, err
	}

	if resp.Tracks.Total > 0 {
		return findSongTrackMatch(resp.Tracks, song)
	}

	return domain.SpotifyTrack{}, errTrackNotFound
}

type match struct {
	track              domain.SpotifyTrack
	artistPercentMatch float64
	trackPercentMatch  float64
	albumPercentMatch  float64
}

func (m match) isExactMatch() bool {
	return m.trackPercentMatch == 100 &&
		m.artistPercentMatch == 100 &&
		m.albumPercentMatch == 100
}

func (m match) weightedAverage() float64 {
	var (
		weightArtist = 0.35
		weightTrack  = 0.40
		weightAlbum  = 0.25
	)

	return m.artistPercentMatch*weightArtist + m.trackPercentMatch*weightTrack + m.albumPercentMatch*weightAlbum
}

func findSongTrackMatch(tracks spotify.TrackCollection, song domain.Song) (domain.SpotifyTrack, error) {
	slog.Info("spotify search tracks found", slog.Int("count", tracks.Total))

	if tracks.Total == 1 {
		t := tracks.Items[0]
		return domain.NewSpotifyTrack(song.ID(), t.ID, t.URI), nil
	}

	var matches []match

	for _, t := range tracks.Items {
		m := match{
			track:              domain.NewSpotifyTrack(song.ID(), t.ID, t.URI),
			trackPercentMatch:  stringSimilarity(song.Track(), t.Name),
			artistPercentMatch: percentArtistMatch(t.Artists, song.Artist()),
			albumPercentMatch:  percentAlbumMatch(t.Album, song),
		}

		if m.isExactMatch() {
			return m.track, nil
		}

		matches = append(matches, m)
	}

	slices.SortFunc(matches, func(a, b match) int {
		if a.weightedAverage() < b.weightedAverage() {
			return 1
		}
		if a.weightedAverage() > b.weightedAverage() {
			return -1
		}
		return 0
	})

	if len(matches) == 0 || matches[0].weightedAverage() < minMatchPercent {
		return domain.SpotifyTrack{}, errMatchBelowThreshold
	}

	slog.Info("partial match track found", slog.Any("percent", matches[0].weightedAverage()))

	return matches[0].track, nil
}

func percentAlbumMatch(trackAlbum spotify.Album, song domain.Song) float64 {
	if trackAlbum.AlbumType == spotify.SingleAlbumType && strings.HasPrefix(song.Album(), song.Track()) {
		return 100.0
	}

	return stringSimilarity(song.Album(), trackAlbum.Name)
}

func percentArtistMatch(artists []spotify.Artist, artist string) float64 {
	if len(artists) == 1 {
		return stringSimilarity(artist, artists[0].Name)
	}

	// TODO: still need to figure out how source handles multiple artists
	// for now just pick the highest?
	var matches []float64
	for _, a := range artists {
		matches = append(matches, stringSimilarity(artist, a.Name))
	}

	slices.Sort(matches)

	return matches[len(matches)-1]
}

func stringSimilarity(s1, s2 string) float64 {
	return compare.StringSimilarity(strings.ToLower(s1), strings.ToLower(s2))
}
