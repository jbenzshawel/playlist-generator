package storage

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jbenzshawel/playlist-generator/internal/domain"
)

func TestRepository(t *testing.T) {
	var (
		now = formatDateTime(t, time.Now())

		datePlayedNowDay = now.Format(time.DateOnly)

		songID1 = uuid.New()
		songID2 = uuid.New()

		songHash1 = "songHash1"
		songHash2 = "songHash2"

		songs = []domain.Song{
			domain.NewSongFromDB(songID1, "artist1", "track1", "album1", "upc1", songHash1, now),
			domain.NewSongFromDB(songID2, "artist2", "track2", "album2", "upc2", songHash2, now),
		}

		songSources = []domain.SongSource{
			domain.NewSongSourceFromDB(uuid.New(), "sourceID1", songHash1, domain.StudioOneSourceType, "Studio One Tracks", datePlayedNowDay, now, now),
			domain.NewSongSourceFromDB(uuid.New(), "sourceID2", songHash2, domain.StudioOneSourceType, "Studio One Tracks", datePlayedNowDay, now, now),
		}

		spotifyTrack = domain.NewSpotifyTrack(songID1, "trackID1", "upc1")

		playlist = domain.NewPlaylistFromDB("id1", "uri1", "2025-01-01", "name1", domain.SpotifyPlaylistType, domain.StudioOneSourceType, "", formatDateTime(t, time.Now()))
	)

	storage := InitTestStorage(t)

	repo := NewRepository(storage)

	t.Run("rollback transaction", func(t *testing.T) {
		require.NoError(t, repo.Begin(t.Context()))

		require.NoError(t, repo.Song().BulkInsert(t.Context(), songs))
		require.NoError(t, repo.SongSource().BulkInsert(t.Context(), songSources))
		require.NoError(t, repo.SpotifyTrack().Insert(t.Context(), spotifyTrack))
		require.NoError(t, repo.Playlist().Insert(t.Context(), playlist))

		require.NoError(t, repo.Rollback())

		assert.Empty(t, getAllSongs(t, storage.db))
		assert.Empty(t, getAllSongSources(t, storage.db))
		assert.Empty(t, getAllPlaylists(t, storage.db))
		assert.Empty(t, getAllSpotifyTracks(t, storage.db))
	})

	t.Run("commit transaction", func(t *testing.T) {
		require.NoError(t, repo.Begin(t.Context()))

		require.NoError(t, repo.Song().BulkInsert(t.Context(), songs))
		require.NoError(t, repo.SongSource().BulkInsert(t.Context(), songSources))
		require.NoError(t, repo.SpotifyTrack().Insert(t.Context(), spotifyTrack))
		require.NoError(t, repo.Playlist().Insert(t.Context(), playlist))

		require.NoError(t, repo.Commit())

		assert.Equal(t, songs, getAllSongs(t, storage.db))
		assert.Equal(t, songSources, getAllSongSources(t, storage.db))
		assert.Equal(t, []domain.Playlist{playlist}, getAllPlaylists(t, storage.db))
		assert.Equal(t, []domain.SpotifyTrack{spotifyTrack}, getAllSpotifyTracks(t, storage.db))
	})
}
