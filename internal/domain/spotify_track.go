package domain

import (
	"context"

	"github.com/google/uuid"
)

type SpotifyTrackRepository interface {
	// GetUnknownSongs returns all songs that haven't been populated with
	// Spotify metadata
	GetUnknownSongs(ctx context.Context) ([]Song, error)

	Insert(ctx context.Context, track SpotifyTrack) error
}

// SpotifyTrack includes the song metadata required to create a spotify playlist.
type SpotifyTrack struct {
	id            string
	uri           string
	songID        uuid.UUID
	matchNotFound bool
}

func NewSpotifyTrack(songID uuid.UUID, trackID, uri string) SpotifyTrack {
	return SpotifyTrack{
		songID: songID,
		id:     trackID,
		uri:    uri,
	}
}

func NewNotFoundSpotifyTrack(songID uuid.UUID) SpotifyTrack {
	return SpotifyTrack{
		matchNotFound: true,
	}
}

func (s SpotifyTrack) TrackID() string {
	return s.id
}

func (s SpotifyTrack) URI() string {
	return s.uri
}

func (s SpotifyTrack) SongID() uuid.UUID {
	return s.songID
}

func (s SpotifyTrack) MatchNotFound() bool {
	return s.matchNotFound
}
