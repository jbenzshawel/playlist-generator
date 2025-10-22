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
	track        domain.SpotifyTrack
	percentMatch float64
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
		percentArtMatch := percentArtistMatch(t.Artists, song.Artist())

		percentAlbumMatch := compare.StringSimilarity(song.Album(), t.Album.Name)

		matches = append(matches, match{
			percentMatch: (percentArtMatch + percentAlbumMatch) / 2,
			track:        domain.NewSpotifyTrack(song.SongHash(), t.ID, t.URI),
		})
	}

	slices.SortFunc(matches, func(a, b match) int {
		if a.percentMatch < b.percentMatch {
			return 1
		}
		if a.percentMatch > b.percentMatch {
			return -1
		}
		return 0
	})

	if len(matches) == 0 {
		return domain.SpotifyTrack{}, errors.New("no matches")
	}

	slog.Info("partial match track found", slog.Any("percent", matches[0].percentMatch))

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

func percentStringMatch(orig, alt string) float64 {
	return compare.StringSimilarity(orig, alt)
}
