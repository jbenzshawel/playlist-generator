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

func TestSongSourceSqlRepository(t *testing.T) {
	t.Parallel()

	var (
		endtime1 = formatDateTime(t, time.Now())
		endtime2 = formatDateTime(t, time.Now().Add(5*time.Minute))
		endtime3 = formatDateTime(t, time.Now().Add(10*time.Minute))
	)

	expectedSongSources := []domain.SongSource{
		domain.NewSongSourceFromDB(uuid.New(), "sourceID1", "songHash1", domain.StudioOneSourceType, "Studio One Tracks", "2025-11-04", endtime1, endtime1),
		domain.NewSongSourceFromDB(uuid.New(), "sourceID2", "songHash2", domain.StudioOneSourceType, "Studio One Tracks", "2025-11-04", endtime2, endtime2),
		domain.NewSongSourceFromDB(uuid.New(), "sourceID3", "songHash3", domain.StudioOneSourceType, "Studio One Tracks", "2025-11-04", endtime3, endtime3),
	}

	storage := InitTestStorage(t)

	r := &songSourceSqlRepository{stmts: storage.stmts}

	tx, err := storage.db.BeginTx(t.Context(), nil)
	require.NoError(t, err)

	r.SetTransaction(tx)

	t.Run("bulk insert", func(t *testing.T) {
		require.NoError(t, r.BulkInsert(t.Context(), expectedSongSources))

		actual := getAllSongSources(t, tx)

		assert.Equal(t, expectedSongSources, actual)
	})

	t.Run("ignores existing songs on insert", func(t *testing.T) {
		require.NoError(t, r.BulkInsert(t.Context(), expectedSongSources))

		actual := getAllSongSources(t, tx)

		assert.Equal(t, expectedSongSources, actual)
	})

	t.Run("commit", func(t *testing.T) {
		require.NoError(t, tx.Commit())
	})
}

func getAllSongSources(t *testing.T, tx *sql.Tx) []domain.SongSource {
	rows, err := tx.QueryContext(
		t.Context(),
		`SELECT * FROM song_sources;`,
	)
	require.NoError(t, err)

	defer func() {
		require.NoError(t, rows.Close())
	}()

	results, err := scanSourceSourceRows(rows)
	require.NoError(t, err)

	return results
}
