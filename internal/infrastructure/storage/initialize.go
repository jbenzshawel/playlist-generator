package storage

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"

	"github.com/jbenzshawel/playlist-generator/internal/infrastructure/storage/internal/statements"
)

var tableSchemas = []string{
	sourceTypeSchema,
	playlistTypeSchema,
	songSchema,
	songSourceSchema,
	spotifyTrackSchema,
	playlistsSchema,
}

var lookupInitializers = []func(*sql.DB) error{
	initSourceTypes,
	initPlaylistTypes,
}

// Initialize opens a database connection and initializes the database. A DB connection
// pool and closer are returned. Note the closer function closes the connection pool for
// you.
func Initialize(ctx context.Context, dsn string) (*sql.DB, func(), error) {
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to open database: %w", err)
	}

	for _, table := range tableSchemas {
		_, err = db.ExecContext(ctx, table)
		if err != nil {
			return nil, nil, fmt.Errorf("error initilizing table schema: %w", err)
		}
	}

	for _, lookupInit := range lookupInitializers {
		err = lookupInit(db)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to seed lookup table: %w", err)
		}
	}

	err = statements.Prepare(ctx, db)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to prepare statements: %w", err)
	}

	closer := func() {
		statements.Close()

		err := db.Close()
		if err != nil {
			slog.Warn("error closing db connection", slog.Any("error", err))
		}
	}

	return db, closer, nil
}
