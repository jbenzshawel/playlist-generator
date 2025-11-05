package storage

import (
	"database/sql"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jbenzshawel/playlist-generator/internal/domain"
)

func TestSongSqlRepository(t *testing.T) {
	t.Parallel()

	now := formatDateTime(t, time.Now())

	expectedSongs := []domain.Song{
		domain.NewSongFromDB(uuid.New(), "artist1", "track1", "album1", "upc1", "songHash1", now),
		domain.NewSongFromDB(uuid.New(), "artist2", "track2", "album2", "upc2", "songHash2", now),
		domain.NewSongFromDB(uuid.New(), "artist3", "track3", "album3", "upc3", "songHash3", now),
	}

	storage := InitTestStorage(t)

	r := &songSqlRepository{stmts: storage.stmts}

	tx, err := storage.db.BeginTx(t.Context(), nil)
	require.NoError(t, err)
	r.SetTransaction(tx)

	t.Run("bulk insert", func(t *testing.T) {
		require.NoError(t, r.BulkInsert(t.Context(), expectedSongs))

		actual := getAllSongs(t, tx)

		assert.Equal(t, expectedSongs, actual)
	})

	t.Run("ignores existing songs on insert", func(t *testing.T) {
		require.NoError(t, r.BulkInsert(t.Context(), expectedSongs))

		actual := getAllSongs(t, tx)

		assert.Equal(t, expectedSongs, actual)
	})

	t.Run("commit", func(t *testing.T) {
		require.NoError(t, tx.Commit())
	})
}

func getAllSongs(t *testing.T, tx *sql.Tx) []domain.Song {
	rows, err := tx.QueryContext(
		t.Context(),
		`SELECT * FROM songs;`,
	)
	require.NoError(t, err)

	defer func() {
		require.NoError(t, rows.Close())
	}()

	results, err := scanSongRows(rows)
	require.NoError(t, err)

	return results
}
