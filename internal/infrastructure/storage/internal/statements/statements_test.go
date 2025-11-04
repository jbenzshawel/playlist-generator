package statements

import (
	"database/sql"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"
)

func TestStatements(t *testing.T) {
	db, err := sql.Open("sqlite", "file::memory:?mode=memory&cache=shared")
	require.NoError(t, err)

	t.Cleanup(func() {
		Close()
		_ = db.Close()
	})

	t.Run("Get when not prepared returns error", func(t *testing.T) {
		_, err = Get(InsertSongType)
		assert.Error(t, err)
	})

	t.Run("Prepare prepares all statements", func(t *testing.T) {
		require.NoError(t, Prepare(t.Context(), db))

		for _, st := range AllTypes() {
			stmt, err := Get(st)
			require.NoError(t, err)
			require.NotNil(t, stmt)
		}
	})
}
