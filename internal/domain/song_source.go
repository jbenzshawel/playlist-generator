package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type SongSourceRepository interface {
	BulkInsert(ctx context.Context, songs []SongSource) error
}

// SongSource represents the source that a playlist's song came from. The source
// is related to a domain.Song through the SongHash to allow creating the relationship
// to a new or existing song without needing to create or look it up ahead of time.
type SongSource struct {
	id          uuid.UUID
	sourceID    string
	songHash    string
	sourceType  SourceType
	programName string
	day         string
	endTime     time.Time
	created     time.Time
}

func NewSongSource(sourceID, songHash string, sourceType SourceType, programName, day string, endTime time.Time) SongSource {
	return SongSource{
		id:          uuid.New(),
		sourceID:    sourceID,
		songHash:    songHash,
		sourceType:  sourceType,
		programName: programName,
		day:         day,
		endTime:     endTime,
		created:     time.Now(),
	}
}

func NewSongSourceFromDB(id uuid.UUID, sourceID, songHash string, sourceType SourceType, programName, day string, endTime, created time.Time) SongSource {
	return SongSource{
		id:          id,
		sourceID:    sourceID,
		songHash:    songHash,
		sourceType:  sourceType,
		programName: programName,
		day:         day,
		endTime:     endTime,
		created:     created,
	}
}

func (s SongSource) ID() uuid.UUID {
	return s.id
}

func (s SongSource) SourceID() string {
	return s.sourceID
}

func (s SongSource) SongHash() string {
	return s.songHash
}

func (s SongSource) SourceType() SourceType {
	return s.sourceType
}

func (s SongSource) ProgramName() string {
	return s.programName
}

func (s SongSource) Day() string {
	return s.day
}

func (s SongSource) EndTime() time.Time {
	return s.endTime
}

func (s SongSource) Created() time.Time {
	return s.created
}
