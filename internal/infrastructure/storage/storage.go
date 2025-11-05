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

type statementGetter interface {
	Get(statements.Type) (*sql.Stmt, error)
	Close()
}

type Storage struct {
	db    *sql.DB
	stmts statementGetter
}

// Initialize opens a database connection and initializes the database. A Storage struct
// is returned with a Close function that closes the database connection.
func Initialize(ctx context.Context, dsn string) (*Storage, error) {
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	for _, table := range tableSchemas {
		_, err = db.ExecContext(ctx, table)
		if err != nil {
			return nil, fmt.Errorf("error initilizing table schema: %w", err)
		}
	}

	for _, lookupInit := range lookupInitializers {
		err = lookupInit(db)
		if err != nil {
			return nil, fmt.Errorf("failed to seed lookup table: %w", err)
		}
	}

	stmts, err := statements.New(ctx, db)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare statements: %w", err)
	}

	return &Storage{db: db, stmts: stmts}, nil
}

func (s *Storage) Close() {
	s.stmts.Close()

	err := s.db.Close()
	if err != nil {
		slog.Warn("error closing database connection", slog.Any("error", err))
	}
}
