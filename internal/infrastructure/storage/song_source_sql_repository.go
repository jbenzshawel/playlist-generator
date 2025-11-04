package storage

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"

	"github.com/jbenzshawel/playlist-generator/internal/domain"
	"github.com/jbenzshawel/playlist-generator/internal/infrastructure/storage/internal/statements"
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
	stmt, err := statements.Get(statements.InsertSongSourceType)
	if err != nil {
		return err
	}

	var insertCount int64
	for _, s := range songs {
		res, err := r.tx.StmtContext(ctx, stmt).
			ExecContext(
				ctx,
				s.ID(), s.SourceID(), s.SourceType(), s.SongHash(), s.ProgramName(), s.Day(), timeToUTCString(s.EndTime()), timeToUTCString(s.Created()),
			)
		if err != nil {
			return fmt.Errorf("failed to insert song source: %w", err)
		}
		count, err := res.RowsAffected()
		if err != nil {
			return fmt.Errorf("failed to get affected rows: %w", err)
		}
		insertCount += count
	}

	slog.Debug("insert song sources complete", slog.Int64("count", insertCount))

	return nil
}
