package storage

import (
	"context"
	"database/sql"
)

func InitializeSchema(ctx context.Context, db *sql.DB) error {
	_, err := db.ExecContext(ctx, songSchema)
	if err != nil {
		return err
	}

	_, err = db.ExecContext(ctx, publicRadioSongSchema)
	if err != nil {
		return err
	}

	return nil
}
