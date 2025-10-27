package storage

import (
	"context"
	"database/sql"

	"github.com/jbenzshawel/playlist-generator/internal/domain"
)

var _ domain.Repository = (*repository)(nil)

type repository struct {
	db *sql.DB
	tx *sql.Tx

	song         *songSqlRepository
	songSource   *songSourceSqlRepository
	spotifyTrack *spotifyTrackSqlRepository
	playlist     *playlistSqlRepository
}

func (r *repository) Song() domain.SongRepository {
	return r.song
}

func (r *repository) SongSource() domain.SongSourceRepository {
	return r.songSource
}

func (r *repository) SpotifyTrack() domain.SpotifyTrackRepository {
	return r.spotifyTrack
}

func (r *repository) Playlist() domain.PlaylistRepository {
	return r.playlist
}

func (r *repository) Begin(ctx context.Context) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	r.tx = tx
	r.song.SetTransaction(tx)
	r.songSource.SetTransaction(tx)
	r.spotifyTrack.SetTransaction(tx)
	r.playlist.SetTransaction(tx)

	return nil
}

func (r *repository) Commit() error {
	return r.tx.Commit()
}

func (r *repository) Rollback() error {
	return r.tx.Rollback()
}

func NewRepository(db *sql.DB) *repository {
	return &repository{
		db:           db,
		song:         &songSqlRepository{},
		songSource:   &songSourceSqlRepository{},
		spotifyTrack: &spotifyTrackSqlRepository{},
		playlist:     &playlistSqlRepository{},
	}
}
