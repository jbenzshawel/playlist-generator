package domain

// PlaylistType represents where the playlist is created.
type PlaylistType int

const (
	UnknownPlaylistType PlaylistType = 0
	SpotifyPlaylistType PlaylistType = 1
)

var playlistTypes = map[PlaylistType]string{
	UnknownPlaylistType: "Unknown",
	SpotifyPlaylistType: "Spotify",
}

func (t PlaylistType) String() string {
	s, ok := playlistTypes[t]
	if !ok {
		return "Unknown"
	}
	return s
}

func (t PlaylistType) IsValid() bool {
	_, ok := playlistTypes[t]
	return ok
}

func AllPlaylistTypes() []PlaylistType {
	return []PlaylistType{
		SpotifyPlaylistType,
	}
}
