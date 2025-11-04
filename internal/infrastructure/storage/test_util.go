//go:build !prod

package storage

import (
	"database/sql"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"
)

func InitTestDB(t *testing.T) *sql.DB {
	t.Helper()

	db, closer, err := Initialize(t.Context(), "file::memory:?mode=memory&cache=shared")
	t.Cleanup(func() {
		closer()
	})
	require.NoError(t, err)

	return db
}

// formatDateTime is a test helper for formating time on an expected struct
// at the precision the time will be when inserted/returned from the database
func formatDateTime(t *testing.T, dt time.Time) time.Time {
	t.Helper()

	dtString := timeToUTCString(dt)
	dt, err := utcStringToTime(dtString)
	require.NoError(t, err)

	return dt
}
