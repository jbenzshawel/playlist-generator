package domain

import (
	"context"
	"log/slog"
	"time"

	"github.com/google/uuid"
)

type SongRepository interface {
	BulkInsert(ctx context.Context, songs []Song) error
}

// Song represents a song downloaded from a playlist data source.
type Song struct {
	id       uuid.UUID
	artist   string
	track    string
	album    string
	upc      string
	songHash string
	created  time.Time
}

func NewSong(artist, track, album, upc string) (Song, error) {
	songHash, err := newSongHash(artist, track, album)
	if err != nil {
		return Song{}, err
	}

	return Song{
		id:       uuid.New(),
		artist:   artist,
		track:    track,
		album:    album,
		upc:      upc,
		songHash: songHash,
		created:  time.Now(),
	}, nil
}

func (s Song) ID() uuid.UUID {
	return s.id
}

func (s Song) Artist() string {
	return s.artist
}

func (s Song) Track() string {
	return s.track
}

func (s Song) Album() string {
	return s.album
}

func (s Song) UPC() string {
	return s.upc
}

// SongHash returns a hash of the artist, track, and album. It is used
// for relationships between various song sources and the domain.Song
func (s Song) SongHash() string {
	return s.songHash
}

func (s Song) Created() time.Time {
	return s.created
}

func (s Song) LogValue() slog.Value {
	return slog.GroupValue(
		slog.String("artist", s.Artist()),
		slog.String("track", s.Track()),
		slog.String("album", s.Album()),
	)
}
