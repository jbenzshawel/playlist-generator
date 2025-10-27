package storage

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/jbenzshawel/playlist-generator/internal/domain"
)

var _ domain.PlaylistRepository = (*playlistSqlRepository)(nil)

var playlistsSchema string = `CREATE TABLE IF NOT EXISTS playlists (
    id TEXT PRIMARY KEY,
    uri TEXT NOT NULL,
    name TEXT NOT NULL,
    date TEXT NOT NULL,
    source_type_id INT NOT NULL,
  	playlist_type_id INT NOT NULL,
    last_day_synced TEXT NOT NULL,        
    created TEXT NOT NULL           -- store timestamps as ISO8601 strings (UTC)
);`

type playlistSqlRepository struct {
	tx *sql.Tx
}

func (r *playlistSqlRepository) SetTransaction(tx *sql.Tx) {
	r.tx = tx
}

func (r *playlistSqlRepository) GetPlaylistByID(ctx context.Context, id string) (domain.Playlist, error) {
	row := r.tx.QueryRowContext(ctx, `SELECT id, uri, name, date, source_type_id, playlist_type_id, last_day_synced, created
		FROM playlists WHERE id = ?`,
		id,
	)

	p, err := r.scanPlaylist(row)
	if err != nil {
		return domain.Playlist{}, err
	}
	return p, nil
}

func (r *playlistSqlRepository) GetPlaylistByDate(ctx context.Context, playlistType domain.PlaylistType, date string) (domain.Playlist, error) {
	row := r.tx.QueryRowContext(ctx, `SELECT id, uri, name, date, source_type_id, playlist_type_id, last_day_synced, created
		FROM playlists WHERE playlist_type_id = ? AND date = ?`,
		playlistType, date,
	)

	p, err := r.scanPlaylist(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Playlist{}, nil
		}
		return domain.Playlist{}, err
	}
	return p, nil
}

func (r *playlistSqlRepository) scanPlaylist(row *sql.Row) (domain.Playlist, error) {
	var (
		id            string
		uri           string
		name          string
		date          string
		playlistType  domain.PlaylistType
		sourceType    domain.SourceType
		lastDaySynced string
		createdStr    string
	)
	err := row.Scan(&id, &uri, &name, &date, &playlistType, &sourceType, &lastDaySynced, &createdStr)

	if err != nil {
		return domain.Playlist{}, err
	}

	created, err := time.Parse(time.RFC3339, createdStr)
	if err != nil {
		return domain.Playlist{}, err
	}

	return domain.NewPlaylistFromDB(id, uri, date, name, playlistType, sourceType, lastDaySynced, created), nil
}

func (r *playlistSqlRepository) Insert(ctx context.Context, playlist domain.Playlist) error {
	_, err := r.tx.ExecContext(
		ctx,
		`INSERT INTO playlists (id, uri, name, date, source_type_id, playlist_type_id, last_day_synced, created)
			VALUES (?,?, ?, ?,?, ?, ?, ?);`,
		playlist.ID(), playlist.URI(), playlist.Name(), playlist.Date(), playlist.SourceType(), playlist.PlaylistType(), "", timeToUTCString(playlist.Created()),
	)
	if err != nil {
		return err
	}

	return nil
}

func (r *playlistSqlRepository) SetLastDaySynced(ctx context.Context, id, lastDaySynced string) error {
	_, err := r.tx.ExecContext(
		ctx,
		`UPDATE platlists SET last_day_synced = ? WHERE id = ?
			VALUES (?,?);`,
		lastDaySynced, id,
	)
	if err != nil {
		return err
	}

	return nil
}
