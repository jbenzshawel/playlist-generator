//go:build !prod

package storage

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"
)

func InitTestStorage(t *testing.T) *Storage {
	t.Helper()

	storage, err := Initialize(t.Context(), "file::memory:?mode=memory&cache=shared")
	t.Cleanup(func() {
		storage.Close()
	})
	require.NoError(t, err)

	return storage
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
