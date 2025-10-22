package domain

import "time"

type SpotifyTrack struct {
	songHash string
	trackID  string
	uri      string
	created  time.Time
}

func NewSpotifyTrack(songHash, trackID, uri string) SpotifyTrack {
	return SpotifyTrack{
		songHash: songHash,
		trackID:  trackID,
		uri:      uri,
		created:  time.Now(),
	}
}

func (s SpotifyTrack) TrackID() string {
	return s.trackID
}

func (s SpotifyTrack) URI() string {
	return s.uri
}

func (s SpotifyTrack) SongHash() string {
	return s.songHash
}

func (s SpotifyTrack) Created() time.Time {
	return s.created
}
