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

	songs         *songSqlRepository
	songSource    *songSourceSqlRepository
	spotifyTracks *spotifyTrackSqlRepository
}

func (r *repository) Songs() domain.SongRepository {
	return r.songs
}

func (r *repository) SongSource() domain.SongSourceRepository {
	return r.songSource
}

func (r *repository) SpotifyTracks() domain.SpotifyTrackRepository {
	return r.spotifyTracks
}

func (r *repository) Begin(ctx context.Context) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	r.tx = tx
	r.songs.SetTransaction(tx)
	r.songSource.SetTransaction(tx)
	r.spotifyTracks.SetTransaction(tx)

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
		db:            db,
		songs:         newSongSqlRepository(),
		songSource:    newSongSourceSqlRepository(),
		spotifyTracks: newSpotifyTracksSqlRepository(),
	}
}
