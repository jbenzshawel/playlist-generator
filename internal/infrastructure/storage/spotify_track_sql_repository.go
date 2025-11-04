package storage

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"

	"github.com/jbenzshawel/playlist-generator/internal/domain"
	"github.com/jbenzshawel/playlist-generator/internal/infrastructure/storage/internal/statements"
)

var _ domain.SpotifyTrackRepository = (*spotifyTrackSqlRepository)(nil)

var spotifyTrackSchema string = `CREATE TABLE IF NOT EXISTS spotify_tracks (
    id TEXT,
    uri TEXT NOT NULL,
    song_id TEXT NOT NULL,
    match_found INTEGER NOT NULL DEFAULT 0 CHECK(match_found IN (0,1)),
    PRIMARY KEY (id, song_id)
);`

type spotifyTrackSqlRepository struct {
	tx *sql.Tx
}

func (r *spotifyTrackSqlRepository) SetTransaction(tx *sql.Tx) {
	r.tx = tx
}

func (r *spotifyTrackSqlRepository) GetUnknownSongs(ctx context.Context) ([]domain.Song, error) {
	rows, err := r.tx.QueryContext(
		ctx,
		`SELECT songs.id, songs.artist, songs.track, songs.album, songs.upc, songs.song_hash, songs.created
			FROM songs 
			LEFT JOIN spotify_tracks ON songs.id = spotify_tracks.song_id
			WHERE spotify_tracks.uri IS NULL;`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []domain.Song
	for rows.Next() {
		var (
			idStr      string
			artist     string
			track      string
			album      string
			upc        sql.NullString
			songHash   string
			createdStr string
		)

		if err := rows.Scan(&idStr, &artist, &track, &album, &upc, &songHash, &createdStr); err != nil {
			return nil, err
		}

		id, err := uuid.Parse(idStr)
		if err != nil {
			return nil, err
		}
		created, err := time.Parse(time.RFC3339, createdStr)
		if err != nil {
			return nil, err
		}

		upcVal := ""
		if upc.Valid {
			upcVal = upc.String
		}

		s := domain.NewSongFromDB(id, artist, track, album, upcVal, songHash, created)
		results = append(results, s)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return results, nil
}

func (r *spotifyTrackSqlRepository) GetTracksPlayedInRange(ctx context.Context, songSourceType domain.SourceType, startDate, endDate string) ([]domain.SpotifyTrack, error) {
	rows, err := r.tx.QueryContext(
		ctx,
		`SELECT DISTINCT spotify_tracks.id, spotify_tracks.uri, spotify_tracks.song_id, spotify_tracks.match_found  
			FROM songs
			JOIN spotify_tracks ON songs.id = spotify_tracks.song_id
			JOIN song_sources ON song_sources.song_hash = songs.song_hash
			WHERE spotify_tracks.match_found = 1
			  AND song_sources.source_type_id = ?
			  AND song_sources.date_played >= ?
			  AND song_sources.date_played < ?`,
		songSourceType, startDate, endDate,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []domain.SpotifyTrack
	for rows.Next() {
		var (
			id            string
			uri           string
			songIDStr     string
			matchFoundInt int
		)

		if err := rows.Scan(&id, &uri, &songIDStr, &matchFoundInt); err != nil {
			return nil, err
		}

		songID, err := uuid.Parse(songIDStr)
		if err != nil {
			return nil, err
		}

		matchFound := false
		if matchFoundInt == 1 {
			matchFound = true
		}

		s := domain.NewSpotifyTrackFromDB(id, uri, songID, matchFound)
		results = append(results, s)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return results, nil
}

func (r *spotifyTrackSqlRepository) Insert(ctx context.Context, track domain.SpotifyTrack) error {
	stmt, err := statements.Get(statements.InsertSpotifyTrackType)
	if err != nil {
		return err
	}

	_, err = r.tx.StmtContext(ctx, stmt).
		ExecContext(
			ctx,
			track.TrackID(), track.URI(), track.SongID(), boolToInt(track.MatchFound()),
		)
	if err != nil {
		return err
	}

	return nil
}
