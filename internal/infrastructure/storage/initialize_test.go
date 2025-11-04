package storage

import (
	"database/sql"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jbenzshawel/playlist-generator/internal/domain"
)

func TestInitializeSchema(t *testing.T) {
	t.Parallel()

	db := InitTestDB(t)

	t.Run("expected tables exists", func(t *testing.T) {
		expectedTables := map[string]struct{}{
			"songs":          {},
			"song_sources":   {},
			"spotify_tracks": {},
			"source_types":   {},
			"playlist_types": {},
			"playlists":      {},
		}

		actualTables := listTables(t, db)
		assert.Equal(t, len(expectedTables), len(actualTables))

		for _, actualTable := range actualTables {
			_, ok := expectedTables[actualTable]
			assert.True(t, ok)
		}
	})

	t.Run("source types lookup", func(t *testing.T) {
		actual := getLookupValues(t, db, "source_types")

		for idx, expected := range domain.AllSourceTypes() {
			assert.Equal(t, expected, domain.SourceType(actual[idx].id))
			assert.Equal(t, expected.String(), actual[idx].name)
		}
	})

	t.Run("playlist types lookup", func(t *testing.T) {
		actual := getLookupValues(t, db, "playlist_types")

		for idx, expected := range domain.AllPlaylistTypes() {
			assert.Equal(t, expected, domain.PlaylistType(actual[idx].id))
			assert.Equal(t, expected.String(), actual[idx].name)
		}
	})
}

func listTables(t *testing.T, db *sql.DB) []string {
	rows, err := db.QueryContext(
		t.Context(),
		`SELECT name FROM sqlite_master WHERE type='table';`,
	)
	require.NoError(t, err)

	defer func() {
		require.NoError(t, rows.Close())
	}()

	var tables []string
	for rows.Next() {
		var name string
		require.NoError(t, rows.Scan(&name))
		tables = append(tables, name)
	}
	require.NoError(t, rows.Err())

	return tables
}

type lookup struct {
	id   int
	name string
}

func getLookupValues(t *testing.T, db *sql.DB, tableName string) []lookup {
	rows, err := db.QueryContext(
		t.Context(),
		fmt.Sprintf(`SELECT id, name FROM %s;`, tableName),
	)
	require.NoError(t, err)

	defer func() {
		require.NoError(t, rows.Close())
	}()

	var lookups []lookup
	for rows.Next() {
		l := lookup{}
		require.NoError(t, rows.Scan(&l.id, &l.name))
		lookups = append(lookups, l)
	}
	require.NoError(t, rows.Err())

	return lookups
}
