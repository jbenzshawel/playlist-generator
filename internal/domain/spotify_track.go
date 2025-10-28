package domain

import (
	"context"

	"github.com/google/uuid"
)

// TODO: rename SpotifyTrack tp Track and add playlist type enum for Spotify?
// I don't plan on adding another playlist location soon, but this change
// would support different playlist locations.

type SpotifyTrackRepository interface {
	// GetUnknownSongs returns all songs that haven't been populated with
	// Spotify metadata
	GetUnknownSongs(ctx context.Context) ([]Song, error)

	// GetTracksPlayedInRange returns the tracks for a source played within a date range. Start is inclusive and end date is exclusive.
	GetTracksPlayedInRange(ctx context.Context, songSourceType SourceType, startDate, endDate string) ([]SpotifyTrack, error)

	Insert(ctx context.Context, track SpotifyTrack) error
}

// SpotifyTrack includes the song metadata required to create a spotify playlist.
type SpotifyTrack struct {
	id         string
	uri        string
	songID     uuid.UUID
	matchFound bool
}

func NewSpotifyTrack(songID uuid.UUID, trackID, uri string) SpotifyTrack {
	return SpotifyTrack{
		songID:     songID,
		id:         trackID,
		uri:        uri,
		matchFound: true,
	}
}

func NewNotFoundSpotifyTrack(songID uuid.UUID) SpotifyTrack {
	return SpotifyTrack{
		songID: songID,
	}
}

func NewSpotifyTrackFromDB(
	id string,
	uri string,
	songID uuid.UUID,
	matchFound bool,
) SpotifyTrack {
	return SpotifyTrack{
		id:         id,
		uri:        uri,
		songID:     songID,
		matchFound: matchFound,
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

func (s SpotifyTrack) MatchFound() bool {
	return s.matchFound
}
