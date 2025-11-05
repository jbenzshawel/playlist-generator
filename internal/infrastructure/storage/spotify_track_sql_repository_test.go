package storage

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jbenzshawel/playlist-generator/internal/domain"
)

func TestSpotifyTrackSqlRepository_GetUnknownSongs(t *testing.T) {
	var (
		now = formatDateTime(t, time.Now())

		songID1 = uuid.New()
		songID2 = uuid.New()
		songID3 = uuid.New()
	)

	storage := InitTestStorage(t)

	tx, err := storage.db.BeginTx(t.Context(), nil)
	require.NoError(t, err)

	songRepo := &songSqlRepository{tx: tx, stmts: storage.stmts}
	trackRepo := &spotifyTrackSqlRepository{tx: tx, stmts: storage.stmts}

	songs := []domain.Song{
		domain.NewSongFromDB(songID1, "artist1", "track1", "album1", "upc1", "songHash1", now),
		domain.NewSongFromDB(songID2, "artist2", "track2", "album2", "upc2", "songHash2", now),
		domain.NewSongFromDB(songID3, "artist3", "track3", "album3", "upc3", "songHash3", now),
	}

	require.NoError(t, songRepo.BulkInsert(t.Context(), songs))

	t.Run("empty spotify tracks", func(t *testing.T) {
		actualSongs, err := trackRepo.GetUnknownSongs(t.Context())
		require.NoError(t, err)

		assert.Equal(t, songs, actualSongs)
	})

	t.Run("two spotify tracks set", func(t *testing.T) {
		require.NoError(t, trackRepo.Insert(t.Context(), domain.NewSpotifyTrack(songID1, "trackID1", "uri1")))
		require.NoError(t, trackRepo.Insert(t.Context(), domain.NewNotFoundSpotifyTrack(songID2)))

		actualSongs, err := trackRepo.GetUnknownSongs(t.Context())
		require.NoError(t, err)

		assert.Equal(t, songs[2:], actualSongs)
	})

	t.Run("all spotify tracks set", func(t *testing.T) {
		require.NoError(t, trackRepo.Insert(t.Context(), domain.NewNotFoundSpotifyTrack(songID3)))

		actualSongs, err := trackRepo.GetUnknownSongs(t.Context())
		require.NoError(t, err)

		assert.Empty(t, actualSongs)
	})
}

func TestSpotifyTrackSqlRepository_GetTracksPlayedInRange(t *testing.T) {
	var (
		now = formatDateTime(t, time.Now())

		datePlayedOldest = formatDateTime(t, time.Now().AddDate(0, 0, -2))
		datePlayedMiddle = formatDateTime(t, time.Now().AddDate(0, 0, -1))
		datePlayedNow    = now

		datePlayedOldestDay = datePlayedOldest.Format(time.DateOnly)
		datePlayedMiddleDay = datePlayedMiddle.Format(time.DateOnly)
		datePlayedNowDay    = datePlayedNow.Format(time.DateOnly)

		songID1 = uuid.New()
		songID2 = uuid.New()
		songID3 = uuid.New()
		songID4 = uuid.New()

		songHash1 = "songHash1"
		songHash2 = "songHash2"
		songHash3 = "songHash3"
		songHash4 = "songHash4"

		songs = []domain.Song{
			domain.NewSongFromDB(songID1, "artist1", "track1", "album1", "upc1", songHash1, now),
			domain.NewSongFromDB(songID2, "artist2", "track2", "album2", "upc2", songHash2, now),
			domain.NewSongFromDB(songID3, "artist3", "track3", "album3", "upc3", songHash3, now),
			domain.NewSongFromDB(songID4, "artist4", "track4", "album4", "upc4", songHash4, now),
		}

		songSources = []domain.SongSource{
			domain.NewSongSourceFromDB(uuid.New(), "sourceID1", songHash1, domain.StudioOneSourceType, "Studio One Tracks", datePlayedOldestDay, datePlayedOldest, datePlayedOldest),
			domain.NewSongSourceFromDB(uuid.New(), "sourceID2", songHash2, domain.StudioOneSourceType, "Studio One Tracks", datePlayedMiddleDay, datePlayedMiddle, datePlayedMiddle),
			domain.NewSongSourceFromDB(uuid.New(), "sourceID3", songHash3, domain.StudioOneSourceType, "Studio One Tracks", datePlayedNowDay, datePlayedNow, datePlayedNow),
			domain.NewSongSourceFromDB(uuid.New(), "sourceID4", songHash4, domain.StudioOneSourceType, "Studio One Tracks", datePlayedNowDay, datePlayedNow, datePlayedNow),
		}

		spotifyTracks = []domain.SpotifyTrack{
			domain.NewSpotifyTrack(songID1, "trackID1", "upc1"),
			domain.NewSpotifyTrack(songID2, "trackID2", "upc2"),
			domain.NewSpotifyTrack(songID3, "trackID3", "upc3"),
		}
	)

	storage := InitTestStorage(t)

	tx, err := storage.db.BeginTx(t.Context(), nil)
	require.NoError(t, err)

	songRepo := &songSqlRepository{tx: tx, stmts: storage.stmts}
	songSourceRepo := &songSourceSqlRepository{tx: tx, stmts: storage.stmts}
	trackRepo := &spotifyTrackSqlRepository{tx: tx, stmts: storage.stmts}

	require.NoError(t, songRepo.BulkInsert(t.Context(), songs))
	require.NoError(t, songSourceRepo.BulkInsert(t.Context(), songSources))

	t.Run("no identified tracks no songs", func(t *testing.T) {
		actual, err := trackRepo.GetTracksPlayedInRange(t.Context(), domain.StudioOneSourceType, datePlayedOldestDay, datePlayedNowDay)
		require.NoError(t, err)
		assert.Empty(t, actual)

		require.NoError(t, trackRepo.Insert(t.Context(), domain.NewNotFoundSpotifyTrack(songID4)))

		actual, err = trackRepo.GetTracksPlayedInRange(t.Context(), domain.StudioOneSourceType, datePlayedOldestDay, datePlayedNowDay)
		require.NoError(t, err)
		assert.Empty(t, actual)
	})

	t.Run("all identified tracks returned", func(t *testing.T) {
		for _, st := range spotifyTracks {
			require.NoError(t, trackRepo.Insert(t.Context(), st))
		}

		exclusiveEndDate := time.Now().AddDate(0, 0, 1).Format(time.DateOnly)

		actual, err := trackRepo.GetTracksPlayedInRange(t.Context(), domain.StudioOneSourceType, datePlayedOldestDay, exclusiveEndDate)
		require.NoError(t, err)
		assert.Equal(t, spotifyTracks[:3], actual)
	})

	t.Run("identified tracks returned for expected date range", func(t *testing.T) {
		actual, err := trackRepo.GetTracksPlayedInRange(t.Context(), domain.StudioOneSourceType, datePlayedOldestDay, datePlayedMiddleDay)
		require.NoError(t, err)
		assert.Equal(t, spotifyTracks[:1], actual)

		actual, err = trackRepo.GetTracksPlayedInRange(t.Context(), domain.StudioOneSourceType, datePlayedMiddleDay, datePlayedNowDay)
		require.NoError(t, err)
		assert.Equal(t, spotifyTracks[1:2], actual)
	})
}

func getAllSpotifyTracks(t *testing.T, db queryContexter) []domain.SpotifyTrack {
	rows, err := db.QueryContext(
		t.Context(),
		`SELECT * FROM spotify_tracks;`,
	)
	require.NoError(t, err)

	defer func() {
		require.NoError(t, rows.Close())
	}()

	results, err := scanSpotifyTracks(rows)
	require.NoError(t, err)

	return results
}
