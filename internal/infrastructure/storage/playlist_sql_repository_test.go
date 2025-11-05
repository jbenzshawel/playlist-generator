package storage

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jbenzshawel/playlist-generator/internal/domain"
)

func TestPlaylistSqlRepository(t *testing.T) {
	t.Parallel()

	const (
		id1 = "id1"
		id2 = "id2"
		id3 = "id3"

		date1 = "2025-01"
		date2 = "2025-01"
		date3 = "2025-02"
	)

	expectedPlaylists := []domain.Playlist{
		domain.NewPlaylistFromDB(id1, "uri1", date1, "name1", domain.SpotifyPlaylistType, domain.StudioOneSourceType, "", formatDateTime(t, time.Now())),
		domain.NewPlaylistFromDB(id2, "uri2", date2, "name2", domain.SpotifyPlaylistType, domain.StudioOneSourceType, "", formatDateTime(t, time.Now().AddDate(0, 0, -1))),
		domain.NewPlaylistFromDB(id3, "uri3", date3, "name3", domain.SpotifyPlaylistType, domain.StudioOneSourceType, "", formatDateTime(t, time.Now())),
	}

	storage := InitTestStorage(t)

	r := &playlistSqlRepository{}

	tx, err := storage.db.BeginTx(t.Context(), nil)
	require.NoError(t, err)
	r.SetTransaction(tx)

	t.Run("insert playlist", func(t *testing.T) {
		for _, expected := range expectedPlaylists {
			require.NoError(t, r.Insert(t.Context(), expected))
		}
	})

	t.Run("get playlist by id", func(t *testing.T) {
		playlist1, err := r.GetPlaylistByID(t.Context(), id1)
		require.NoError(t, err)
		assert.Equal(t, expectedPlaylists[0], playlist1)

		playlist2, err := r.GetPlaylistByID(t.Context(), id2)
		require.NoError(t, err)
		assert.Equal(t, expectedPlaylists[1], playlist2)

		playlist3, err := r.GetPlaylistByID(t.Context(), id3)
		require.NoError(t, err)
		assert.Equal(t, expectedPlaylists[2], playlist3)
	})

	t.Run("get playlist by date", func(t *testing.T) {
		// Playlist 1 and 2 have same date so newest should be returned (playlist 1)
		playlist1, err := r.GetPlaylistByDate(t.Context(), domain.SpotifyPlaylistType, expectedPlaylists[0].Date())
		require.NoError(t, err)
		assert.Equal(t, expectedPlaylists[0], playlist1)

		playlist3, err := r.GetPlaylistByDate(t.Context(), domain.SpotifyPlaylistType, expectedPlaylists[2].Date())
		require.NoError(t, err)
		assert.Equal(t, expectedPlaylists[2], playlist3)
	})

	t.Run("set last synced date", func(t *testing.T) {
		lastSync := "2025-11-04"
		require.NoError(t, r.SetLastDaySynced(t.Context(), id1, lastSync))

		playlist, err := r.GetPlaylistByID(t.Context(), id1)
		require.NoError(t, err)
		assert.Equal(t, lastSync, playlist.LastDaySynced())
	})

	t.Run("commit", func(t *testing.T) {
		require.NoError(t, tx.Commit())
	})
}
