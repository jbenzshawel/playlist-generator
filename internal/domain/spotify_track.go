package domain

import "github.com/google/uuid"

// SpotifyTrack includes the song metadata required to create a spotify playlist.
type SpotifyTrack struct {
	songID  uuid.UUID
	trackID string
	uri     string
}

func NewSpotifyTrack(songID uuid.UUID, trackID, uri string) SpotifyTrack {
	return SpotifyTrack{
		songID:  songID,
		trackID: trackID,
		uri:     uri,
	}
}

func (s SpotifyTrack) TrackID() string {
	return s.trackID
}

func (s SpotifyTrack) URI() string {
	return s.uri
}

func (s SpotifyTrack) SongID() uuid.UUID {
	return s.songID
}
