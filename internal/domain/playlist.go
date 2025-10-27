package domain

import (
	"context"
	"log/slog"
	"time"
)

type PlaylistRepository interface {
	GetPlaylistByID(ctx context.Context, id string) (Playlist, error)
	// GetPlaylistByDate returns a playlist that matches a type and date. Note playlists could be scoped
	// to month, day, or year so date is intentionally a string. Follow YYYY-MM-DD convention.
	GetPlaylistByDate(ctx context.Context, playlistType PlaylistType, date string) (Playlist, error)

	Insert(ctx context.Context, playlist Playlist) error
	SetLastDaySynced(ctx context.Context, id, lastDaySynced string) error
}

// Playlist represents a playlist created from the generator.
type Playlist struct {
	id            string
	uri           string
	name          string
	date          string
	playlistType  PlaylistType
	sourceType    SourceType
	lastDaySynced string
	created       time.Time
}

func NewPlaylist(id, uri, name, date string, playlistType PlaylistType, sourceType SourceType) Playlist {
	return Playlist{
		id:           id,
		uri:          uri,
		name:         name,
		date:         date,
		playlistType: playlistType,
		sourceType:   sourceType,
		created:      time.Now(),
	}
}

func NewPlaylistFromDB(id, uri, date, name string, playlistType PlaylistType, sourceType SourceType, lastDaySynced string, created time.Time) Playlist {
	return Playlist{
		id:            id,
		uri:           uri,
		name:          name,
		date:          date,
		playlistType:  playlistType,
		sourceType:    sourceType,
		lastDaySynced: lastDaySynced,
		created:       created,
	}
}
func (p Playlist) IsZero() bool {
	return p == Playlist{}
}

func (p Playlist) ID() string {
	return p.id
}

func (p Playlist) URI() string {
	return p.uri
}

func (p Playlist) Name() string {
	return p.name
}

func (p Playlist) Date() string {
	return p.date
}

func (p Playlist) PlaylistType() PlaylistType {
	return p.playlistType
}

func (p Playlist) SourceType() SourceType {
	return p.sourceType
}

func (p Playlist) LastDaySynced() string {
	return p.lastDaySynced
}

func (p Playlist) Created() time.Time {
	return p.created
}

func (s Playlist) LogValue() slog.Value {
	return slog.GroupValue(
		slog.String("date", s.Date()),
		slog.String("lastDaySynced", s.LastDaySynced()),
		slog.String("name", s.Name()),
		slog.String("uri", s.URI()),
	)
}
