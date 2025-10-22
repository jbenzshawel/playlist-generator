package domain

import (
	"context"
	"time"
)

type PublicRadioRepository interface {
	BulkInsert(ctx context.Context, songs []PublicRadioSong) error
}

type PublicRadioSong struct {
	id          string
	songHash    string
	programName string
	day         string
	endTime     time.Time
	created     time.Time
}

func NewPublicRadioSong(id, songHash, programName, day string, endTime time.Time) PublicRadioSong {
	return PublicRadioSong{
		id:          id,
		songHash:    songHash,
		programName: programName,
		day:         day,
		endTime:     endTime,
		created:     time.Now(),
	}
}

func (p PublicRadioSong) ID() string {
	return p.id
}

func (p PublicRadioSong) SongHash() string {
	return p.songHash
}

func (p PublicRadioSong) ProgramName() string {
	return p.programName
}

func (p PublicRadioSong) Day() string {
	return p.day
}

func (p PublicRadioSong) EndTime() time.Time {
	return p.endTime
}

func (p PublicRadioSong) Created() time.Time {
	return p.created
}
