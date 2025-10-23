package domain

type SpotifyTrack struct {
	songHash string
	trackID  string
	uri      string
}

func NewSpotifyTrack(songHash, trackID, uri string) SpotifyTrack {
	return SpotifyTrack{
		songHash: songHash,
		trackID:  trackID,
		uri:      uri,
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
