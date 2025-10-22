package spotify

import (
	"context"
	"errors"
	"log/slog"
	"slices"

	"github.com/jbenzshawel/playlist-generator/internal/common/compare"
	"github.com/jbenzshawel/playlist-generator/internal/domain"
)

type trackSearcher interface {
	SearchTrack(ctx context.Context, artist, track, album string) (SearchTrackResponse, error)
}

type spotifyTrackProvider struct {
	searcher trackSearcher
}

func (s spotifyTrackProvider) GetTrack(ctx context.Context, song domain.Song) (domain.SpotifyTrack, error) {
	// TODO: It'd be nice if we also populated ISRC info from spotify in the song db

	resp, err := s.searcher.SearchTrack(ctx, song.Track(), song.Artist(), song.Album())
	if err != nil {
		return domain.SpotifyTrack{}, err
	}

	if resp.Tracks.Total > 0 {
		return findTrackFromResult(resp.Tracks, song)
	}

	slog.Info("track not found with album query param; searching without album", slog.Any("song", song))

	resp, err = s.searcher.SearchTrack(ctx, song.Track(), song.Artist(), "")
	if err != nil {
		return domain.SpotifyTrack{}, err
	}

	if resp.Tracks.Total > 0 {
		return findTrackFromResult(resp.Tracks, song)
	}

	return domain.SpotifyTrack{}, errors.New("track not found") // TODO: custom error?
}

type match struct {
	track              domain.SpotifyTrack
	artistPercentMatch float64
	trackPercentMatch  float64
	albumPercentMatch  float64
}

func (m match) weightedAverage() float64 {
	var (
		weightArtist = 0.35
		weightTrack  = 0.40
		weightAlbum  = 0.25
	)

	return m.artistPercentMatch*weightArtist + m.trackPercentMatch*weightTrack + m.albumPercentMatch*weightAlbum
}

func findTrackFromResult(tracks TrackCollection, song domain.Song) (domain.SpotifyTrack, error) {
	slog.Info("spotify search tracks found", slog.Int("count", tracks.Total))

	if tracks.Total == 1 {
		// Wohoo! Exact match
		t := tracks.Items[0]
		return domain.NewSpotifyTrack(song.SongHash(), t.ID, t.URI), nil
	}

	var matches []match

	for _, t := range tracks.Items {

		matches = append(matches, match{
			track:              domain.NewSpotifyTrack(song.SongHash(), t.ID, t.URI),
			trackPercentMatch:  compare.StringSimilarity(song.Track(), t.Name),
			artistPercentMatch: percentArtistMatch(t.Artists, song.Artist()),
			albumPercentMatch:  compare.StringSimilarity(song.Album(), t.Album.Name),
		})
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

	if len(matches) == 0 || matches[0].weightedAverage() < 60 {
		return domain.SpotifyTrack{}, errors.New("no matches")
	}

	slog.Info("partial match track found", slog.Any("percent", matches[0].weightedAverage()))

	return matches[0].track, nil
}

func percentArtistMatch(artists []Artist, artist string) float64 {
	if len(artists) == 1 {
		return compare.StringSimilarity(artist, artists[0].Name)
	}

	// still need to figure out how source handles multiple artists
	// for now just pick the highest?
	var matches []float64
	for _, a := range artists {
		matches = append(matches, compare.StringSimilarity(artist, a.Name))
	}

	slices.Sort(matches)

	return matches[len(matches)-1]
}
