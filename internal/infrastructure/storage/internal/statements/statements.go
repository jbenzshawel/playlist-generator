package statements

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
)

type Type int

const (
	UnknownType Type = iota
	InsertSongType
	InsertSongSourceType
	InsertSpotifyTrackType
)

var types = map[Type]string{
	InsertSongType:         "InsertSongType",
	InsertSongSourceType:   "InsertSongSourceType",
	InsertSpotifyTrackType: "InsertSpotifyTrackType",
}

func (t Type) String() string {
	s, ok := types[t]
	if !ok {
		return "UnknownType"
	}
	return s
}

func AllTypes() []Type {
	return []Type{
		InsertSongType,
		InsertSongSourceType,
		InsertSpotifyTrackType,
	}
}

const (
	insertSongSQL = `INSERT INTO songs (id, artist, track, album, upc, song_hash, created) 
					VALUES (?,?,?, ?, ?, ?, ?)
					ON CONFLICT (song_hash) DO NOTHING;`

	insertSongSourceSQL = `INSERT INTO song_sources (id, source_id, source_type_id, song_hash, program_name, date_played, end_time, created)
			VALUES (?,?,?,?,?,?,?, ?);`

	insertSpotifyTrackSQL = `INSERT INTO spotify_tracks (id, uri, song_id, match_found)
			VALUES (?,?,?,?);`
)

type statements struct {
	prepared map[Type]*sql.Stmt
}

func New(ctx context.Context, db *sql.DB) (*statements, error) {
	s := &statements{
		prepared: make(map[Type]*sql.Stmt),
	}

	statementsToPrepare := []struct {
		Type Type
		SQL  string
	}{
		{InsertSongType, insertSongSQL},
		{InsertSongSourceType, insertSongSourceSQL},
		{InsertSpotifyTrackType, insertSpotifyTrackSQL},
	}
	for _, stmt := range statementsToPrepare {
		var err error
		s.prepared[stmt.Type], err = db.PrepareContext(ctx, stmt.SQL)
		if err != nil {
			return nil, err
		}
	}

	return s, nil
}

func (s *statements) Get(st Type) (*sql.Stmt, error) {
	stmt, ok := s.prepared[st]
	if !ok {
		return nil, fmt.Errorf("statement %s not found", st.String())
	}
	return stmt, nil
}

func (s *statements) Close() {
	for _, stmt := range s.prepared {
		err := stmt.Close()
		if err != nil {
			slog.Warn("error closing prepared statement", slog.Any("error", err))
		}
	}
}
