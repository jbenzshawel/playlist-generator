package storage

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"

	"github.com/jbenzshawel/playlist-generator/internal/domain"
)

var _ domain.SpotifyTrackRepository = (*spotifyTrackSqlRepository)(nil)

// TODO: Delete match not found?

var spotifyTrackSchema string = `CREATE TABLE IF NOT EXISTS spotify_tracks (
    id TEXT,
    uri TEXT NOT NULL,
    song_id TEXT NOT NULL,
    match_not_found INTEGER NOT NULL DEFAULT 0 CHECK(match_not_found IN (0,1)),
    PRIMARY KEY (id, song_id)
);`

type spotifyTrackSqlRepository struct {
	tx *sql.Tx
}

func newSpotifyTracksSqlRepository() *spotifyTrackSqlRepository {
	return &spotifyTrackSqlRepository{}
}

func (r *spotifyTrackSqlRepository) SetTransaction(tx *sql.Tx) {
	r.tx = tx
}

func (r *spotifyTrackSqlRepository) GetUnknownSongs(ctx context.Context) ([]domain.Song, error) {
	rows, err := r.tx.QueryContext(
		ctx,
		`SELECT songs.* 
		FROM songs LEFT JOIN spotify_tracks ON songs.id = spotify_tracks.song_id
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

func (r spotifyTrackSqlRepository) Insert(ctx context.Context, track domain.SpotifyTrack) error {
	_, err := r.tx.ExecContext(
		ctx,
		`INSERT INTO spotify_tracks (id, uri, song_id, match_not_found)
			VALUES (?,?,?,?);`,
		track.TrackID(), track.URI(), track.SongID(), boolToInt(track.MatchNotFound()),
	)
	if err != nil {
		return err
	}

	return nil
}
