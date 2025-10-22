package domain

import "time"

type SpotifyPlaylist struct {
	id      string
	name    string
	created time.Time
}

func NewSpotifyPlaylist(id, name string) SpotifyPlaylist {
	return SpotifyPlaylist{
		id:      id,
		name:    name,
		created: time.Now(),
	}
}

func (p SpotifyPlaylist) ID() string {
	return p.id
}

func (p SpotifyPlaylist) Name() string {
	return p.name
}

func (p SpotifyPlaylist) Created() time.Time {
	return p.created
}
