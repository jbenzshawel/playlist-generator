package domain

import (
	"context"
	"time"
)

type StudioOneRepository interface {
	BulkInsert(ctx context.Context, songs []StudioOneSource) error
}

// StudioOneSource represents the source that a playlist's song came from. The source
// is related to a domain.Song through the SongHash to allow creating the relationship
// to a new or existing song without needing to create or look it up ahead of time.
type StudioOneSource struct {
	id          string
	songHash    string
	programName string
	day         string
	endTime     time.Time
	created     time.Time
}

func NewStudioOneSource(id, songHash, programName, day string, endTime time.Time) StudioOneSource {
	return StudioOneSource{
		id:          id,
		songHash:    songHash,
		programName: programName,
		day:         day,
		endTime:     endTime,
		created:     time.Now(),
	}
}

func (p StudioOneSource) ID() string {
	return p.id
}

func (p StudioOneSource) SongHash() string {
	return p.songHash
}

func (p StudioOneSource) ProgramName() string {
	return p.programName
}

func (p StudioOneSource) Day() string {
	return p.day
}

func (p StudioOneSource) EndTime() time.Time {
	return p.endTime
}

func (p StudioOneSource) Created() time.Time {
	return p.created
}
