package statements

import (
	"database/sql"
	"testing"

	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"
)

func TestStatements(t *testing.T) {
	db, err := sql.Open("sqlite", "file::memory:?mode=memory&cache=shared")
	require.NoError(t, err)

	s, err := New(t.Context(), db)
	require.NoError(t, err)
	t.Cleanup(func() {
		s.Close()
		_ = db.Close()
	})

	t.Run("Prepare prepares all statements", func(t *testing.T) {
		for _, st := range AllTypes() {
			stmt, err := s.Get(st)
			require.NoError(t, err)
			require.NotNil(t, stmt)
		}
	})
}
