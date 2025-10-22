package domain

import (
	"time"

	"github.com/google/uuid"
)

type SongRepository interface {
	BulkInsert(song []Song) error
}

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
	songHash, err := NewSongHash(artist, track, album)
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

func (s Song) IsZero() bool {
	return s == Song{}
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

func (s Song) SongHash() string {
	return s.songHash
}

func (s Song) Created() time.Time {
	return s.created
}
