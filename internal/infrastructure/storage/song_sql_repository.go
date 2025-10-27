package storage

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jbenzshawel/playlist-generator/internal/domain"
)

var _ domain.SongRepository = (*songSqlRepository)(nil)

var songSchema string = `CREATE TABLE IF NOT EXISTS song (
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
	for _, s := range songs {
		_, err := r.tx.ExecContext(
			ctx,
			`INSERT INTO song (id, artist, track, album, upc, song_hash, created) 
					VALUES (?,?,?, ?, ?, ?, ?)
					ON CONFLICT(song_hash) DO NOTHING;`,
			s.ID(), s.Artist(), s.Track(), s.Album(), s.UPC(), s.SongHash(), timeToUTCString(s.Created()),
		)
		if err != nil {
			return fmt.Errorf("failed to insert song: %w", err)
		}
	}

	return nil
}
