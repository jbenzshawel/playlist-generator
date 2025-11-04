package storage

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"

	"github.com/jbenzshawel/playlist-generator/internal/domain"
	"github.com/jbenzshawel/playlist-generator/internal/infrastructure/storage/internal/statements"
)

var _ domain.SongRepository = (*songSqlRepository)(nil)

var songSchema string = `CREATE TABLE IF NOT EXISTS songs (
    id TEXT PRIMARY KEY,            -- UUID string (e.g., 550e8400-e29b-41d4-a716-446655440000)
    artist TEXT NOT NULL,
    track TEXT NOT NULL,
    album TEXT NOT NULL,
    upc TEXT,
    song_hash TEXT NOT NULL UNIQUE, -- derived hash to de-duplicate song
    created TEXT NOT NULL           -- store timestamps as ISO8601 strings (UTC)
);`

type songSqlRepository struct {
	tx *sql.Tx
}

func (r *songSqlRepository) SetTransaction(tx *sql.Tx) {
	r.tx = tx
}

func (r *songSqlRepository) BulkInsert(ctx context.Context, songs []domain.Song) error {
	stmt, err := statements.Get(statements.InsertSongType)
	if err != nil {
		return err
	}

	var insertCount int64
	for _, s := range songs {
		res, err := r.tx.StmtContext(ctx, stmt).
			ExecContext(
				ctx,
				s.ID(), s.Artist(), s.Track(), s.Album(), s.UPC(), s.SongHash(), timeToUTCString(s.Created()),
			)
		if err != nil {
			return fmt.Errorf("failed to insert song: %w", err)
		}
		count, err := res.RowsAffected()
		if err != nil {
			return fmt.Errorf("failed to get affected rows: %w", err)
		}
		insertCount += count
	}

	slog.Debug("insert songs complete", slog.Int64("count", insertCount))

	return nil
}
