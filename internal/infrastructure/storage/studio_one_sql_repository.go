package storage

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jbenzshawel/playlist-generator/internal/domain"
)

var _ domain.StudioOneRepository = (*StudioOneSqlRepository)(nil)

var studioOneSongSchema string = `CREATE TABLE IF NOT EXISTS studio_one_source (
    id TEXT PRIMARY KEY,
    song_hash TEXT NOT NULL,
    programName TEXT,
    datePlayed TEXT NOT NULL,
    end_time TEXT NOT NULL,         -- store timestamps as ISO8601 strings (UTC)
    created TEXT NOT NULL           -- store timestamps as ISO8601 strings (UTC)
);`

type StudioOneSqlRepository struct {
	db *sql.DB
}

func NewStudioOneSqlRepository(db *sql.DB) *StudioOneSqlRepository {
	return &StudioOneSqlRepository{
		db: db,
	}
}

func (r StudioOneSqlRepository) BulkInsert(ctx context.Context, songs []domain.StudioOneSource) error {
	for _, s := range songs {
		_, err := r.db.ExecContext(
			ctx,
			`INSERT INTO studio_one_source (id, song_hash, programName, datePlayed, end_time, created)
			VALUES (?,?,?,?,?,?)
			ON CONFLICT(song_hash) DO NOTHING;`,
			s.ID(), s.SongHash(), s.ProgramName(), s.Day(), timeToUTCString(s.EndTime()), timeToUTCString(s.Created()),
		)
		if err != nil {
			return fmt.Errorf("failed to insert studio one song: %w", err)
		}
	}

	return nil
}
