package domain

import "time"

type SpotifySong struct {
	trackID  string
	songHash string
	created  time.Time
}

func NewSpotifySong(trackID, songHash string) SpotifySong {
	return SpotifySong{
		trackID:  trackID,
		songHash: songHash,
		created:  time.Now(),
	}
}

func (s SpotifySong) TrackID() string {
	return s.trackID
}

func (s SpotifySong) SongHash() string {
	return s.songHash
}

func (s SpotifySong) Created() time.Time {
	return s.created
}
