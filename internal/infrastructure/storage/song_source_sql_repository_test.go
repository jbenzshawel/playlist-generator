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

	db := InitTestDB(t)

	var (
		endtime1 = formatDateTime(t, time.Now())
		endtime2 = formatDateTime(t, time.Now().Add(5*time.Minute))
		endtime3 = formatDateTime(t, time.Now().Add(10*time.Minute))
	)

	const (
		songHash1 = "songHash1"
		songHash2 = "songHash2"
		songHash3 = "songHash3"
	)

	expectedSongSources := []domain.SongSource{
		domain.NewSongSourceFromDB(uuid.New(), "sourceID1", songHash1, domain.StudioOneSourceType, "Studio One Tracks", "2025-11-04", endtime1, endtime1),
		domain.NewSongSourceFromDB(uuid.New(), "sourceID2", songHash2, domain.StudioOneSourceType, "Studio One Tracks", "2025-11-04", endtime2, endtime2),
		domain.NewSongSourceFromDB(uuid.New(), "sourceID3", songHash3, domain.StudioOneSourceType, "Studio One Tracks", "2025-11-04", endtime3, endtime3),
	}

	r := &songSourceSqlRepository{}

	tx, err := db.BeginTx(t.Context(), nil)
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

	var results []domain.SongSource
	for rows.Next() {
		var (
			idStr       string
			sourceID    string
			songHash    string
			sourceType  domain.SourceType
			programName string
			day         string
			endTimeStr  string
			createdStr  string
		)

		err := rows.Scan(&idStr, &sourceID, &songHash, &sourceType, &programName, &day, &endTimeStr, &createdStr)
		require.NoError(t, err)

		id, err := uuid.Parse(idStr)
		require.NoError(t, err)

		endTime, err := utcStringToTime(endTimeStr)
		require.NoError(t, err)

		created, err := utcStringToTime(createdStr)
		require.NoError(t, err)

		s := domain.NewSongSourceFromDB(id, sourceID, songHash, sourceType, programName, day, endTime, created)
		results = append(results, s)
	}

	require.NoError(t, rows.Err())

	return results
}
