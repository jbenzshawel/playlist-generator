package storage

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	
	"github.com/jbenzshawel/playlist-generator/internal/domain"
)

var _ domain.SpotifyTrackRepository = (*SpotifyTrackSqlRepository)(nil)

var spotifyTrackSchema string = `CREATE TABLE IF NOT EXISTS spotify_tracks (
    id TEXT PRIMARY KEY,
    uri TEXT NOT NULL,
    songId TEXT NOT NULL,
    matchNotFound INTEGER NOT NULL DEFAULT 0 CHECK(matchNotFound IN (0,1))
);`

type SpotifyTrackSqlRepository struct {
	db *sql.DB
}

func NewSpotifyTracksSqlRepository(db *sql.DB) *SpotifyTrackSqlRepository {
	return &SpotifyTrackSqlRepository{
		db: db,
	}
}

func (r SpotifyTrackSqlRepository) GetUnknownSongs(ctx context.Context) ([]domain.Song, error) {
	rows, err := r.db.QueryContext(
		ctx,
		`SELECT songs.* 
		FROM songs LEFT JOIN spotify_tracks ON songs.id = spotify_tracks.songId
		WHERE coalesce(spotify_tracks.matchNotFound, 0) = 0;`,
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

func (r SpotifyTrackSqlRepository) Insert(ctx context.Context, track domain.SpotifyTrack) error {
	_, err := r.db.ExecContext(
		ctx,
		`INSERT INTO spotify_tracks (id, uri, songID, matchNotFound)
			VALUES (?,?,?,?);`,
		track.TrackID(), track.URI(), track.SongID(), boolToInt(track.MatchNotFound()),
	)
	if err != nil {
		return err
	}

	return nil
}
