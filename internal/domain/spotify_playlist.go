package domain

import (
	"context"
	"time"
)

type SpotifyPlaylistRepository interface {
	Insert(ctx context.Context, playlist SpotifyPlaylist) error
	SetLastDaySynced(ctx context.Context, id, lastDaySynced string) error
}

// SpotifyPlaylist represents a playlist created from the generator.
type SpotifyPlaylist struct {
	id            string
	uri           string
	name          string
	sourceType    SourceType
	lastDaySynced string
	created       time.Time
}

func NewSpotifyPlaylist(id, name string, sourceType SourceType, lastDaySynced string) SpotifyPlaylist {
	return SpotifyPlaylist{
		id:            id,
		name:          name,
		sourceType:    sourceType,
		lastDaySynced: lastDaySynced,
		created:       time.Now(),
	}
}

func (p SpotifyPlaylist) ID() string {
	return p.id
}

func (p SpotifyPlaylist) URI() string {
	return p.uri
}

func (p SpotifyPlaylist) Name() string {
	return p.name
}

func (p SpotifyPlaylist) SourceType() SourceType {
	return p.sourceType
}

func (p SpotifyPlaylist) LastDaySynced() string {
	return p.lastDaySynced
}

func (p SpotifyPlaylist) Created() time.Time {
	return p.created
}
