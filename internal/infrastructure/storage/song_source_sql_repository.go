package storage

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jbenzshawel/playlist-generator/internal/domain"
)

var _ domain.SongSourceRepository = (*songSourceSqlRepository)(nil)

var songSourceSchema string = `CREATE TABLE IF NOT EXISTS song_sources (
    id TEXT PRIMARY KEY,
    source_id TEXT NOT NULL,
    song_hash TEXT NOT NULL,
    source_type_id INT NOT NULL,
    program_name TEXT,
    date_played TEXT NOT NULL,
    end_time TEXT NOT NULL,         -- store timestamps as ISO8601 strings (UTC)
    created TEXT NOT NULL,           -- store timestamps as ISO8601 strings (UTC)
   UNIQUE(source_id, source_type_id, end_time) ON CONFLICT IGNORE
);`

type songSourceSqlRepository struct {
	tx *sql.Tx
}

func (r *songSourceSqlRepository) SetTransaction(tx *sql.Tx) {
	r.tx = tx
}

func (r *songSourceSqlRepository) BulkInsert(ctx context.Context, songs []domain.SongSource) error {
	for _, s := range songs {
		_, err := r.tx.ExecContext(
			ctx,
			`INSERT INTO song_sources (id, source_id, source_type_id, song_hash, program_name, date_played, end_time, created)
			VALUES (?,?,?,?,?,?,?, ?);`,
			s.ID(), s.SourceID(), s.SourceType(), s.SongHash(), s.ProgramName(), s.Day(), timeToUTCString(s.EndTime()), timeToUTCString(s.Created()),
		)
		if err != nil {
			return fmt.Errorf("failed to insert song source: %w", err)
		}
	}

	return nil
}
