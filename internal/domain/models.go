package domain

import "time"

type Song struct {
	ID      string
	Artist  string // TODO: Create separate artists table? Not needed for original idea
	Track   string
	Album   string
	UPC     string
	ISRC    string
	Created time.Time
}

type SpotifySong struct {
	TrackID string
	SongID  string
	Created time.Time
}

type SpotifyPlaylist struct {
	ID      string
	Name    string
	Created time.Time
}

type PublicRadioSong struct {
	ID          string
	SongID      string
	ProgramName string
	Day         string
	EndTime     time.Time
	Created     time.Time
}
