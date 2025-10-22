package storage

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jbenzshawel/playlist-generator/internal/domain"
)

var _ domain.PublicRadioRepository = (*PublicRadioSqlRepository)(nil)

var publicRadioSongSchema string = `CREATE TABLE IF NOT EXISTS public_radio_songs (
    id TEXT PRIMARY KEY,
    song_hash TEXT NOT NULL UNIQUE,
    programName TEXT,
    datePlayed TEXT NOT NULL,
    end_time TEXT NOT NULL,         -- store timestamps as ISO8601 strings (UTC)
    created TEXT NOT NULL           -- store timestamps as ISO8601 strings (UTC)
);`

type PublicRadioSqlRepository struct {
	db *sql.DB
}

func NewPublicRadioSqlRepository(db *sql.DB) *PublicRadioSqlRepository {
	return &PublicRadioSqlRepository{
		db: db,
	}
}

func (r PublicRadioSqlRepository) BulkInsert(ctx context.Context, songs []domain.PublicRadioSong) error {
	for _, s := range songs {
		_, err := r.db.ExecContext(
			ctx,
			`INSERT INTO public_radio_songs (id, song_hash, programName, datePlayed, end_time, created)
			VALUES (?,?,?,?,?,?)
			ON CONFLICT(song_hash) DO NOTHING;`,
			s.ID(), s.SongHash(), s.ProgramName(), s.Day(), timeToUTCString(s.EndTime()), timeToUTCString(s.Created()),
		)
		if err != nil {
			return fmt.Errorf("failed to insert public radio song: %w", err)
		}
	}

	return nil
}
