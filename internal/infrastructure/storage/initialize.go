package storage

import (
	"context"
	"database/sql"
)

func InitializeSchema(ctx context.Context, db *sql.DB) error {
	_, err := db.ExecContext(ctx, sourceTypeSchema)
	if err != nil {
		return err
	}

	err = initSourceTypes(db)
	if err != nil {
		return err
	}

	_, err = db.ExecContext(ctx, songSchema)
	if err != nil {
		return err
	}

	_, err = db.ExecContext(ctx, songSourceSchema)
	if err != nil {
		return err
	}

	_, err = db.ExecContext(ctx, spotifyTrackSchema)
	if err != nil {
		return err
	}

	return nil
}
