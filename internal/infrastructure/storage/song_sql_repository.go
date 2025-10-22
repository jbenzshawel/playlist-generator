package storage

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jbenzshawel/playlist-generator/internal/domain"
)

var _ domain.SongRepository = (*SongSqlRepository)(nil)

var songSchema string = `CREATE TABLE IF NOT EXISTS songs (
    id TEXT PRIMARY KEY,            -- UUID string (e.g., 550e8400-e29b-41d4-a716-446655440000)
    artist TEXT NOT NULL,
    track TEXT NOT NULL,
    album TEXT NOT NULL,
    upc TEXT,
    song_hash TEXT NOT NULL UNIQUE, -- derived hash to de-duplicate songs
    created TEXT NOT NULL           -- store timestamps as ISO8601 strings (UTC)
);`

type SongSqlRepository struct {
	db *sql.DB
}

func NewSongSqlRepository(db *sql.DB) *SongSqlRepository {
	return &SongSqlRepository{
		db: db,
	}
}

func (r SongSqlRepository) BulkInsert(ctx context.Context, songs []domain.Song) error {
	for _, s := range songs {
		_, err := r.db.ExecContext(
			ctx,
			`INSERT INTO songs (id, artist, track, album, upc, song_hash, created) 
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
