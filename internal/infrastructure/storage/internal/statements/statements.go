package statements

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"sync"
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

var (
	prepared = map[Type]*sql.Stmt{}

	syncOnce = sync.Once{}
)

func Prepare(ctx context.Context, db *sql.DB) error {
	var syncErr error

	statementsToPrepare := []struct {
		Type Type
		SQL  string
	}{
		{InsertSongType, insertSongSQL},
		{InsertSongSourceType, insertSongSourceSQL},
		{InsertSpotifyTrackType, insertSpotifyTrackSQL},
	}

	syncOnce.Do(func() {
		var err error
		for _, s := range statementsToPrepare {
			prepared[s.Type], err = db.PrepareContext(ctx, s.SQL)
			if err != nil {
				syncErr = err
				return
			}
		}
	})

	return syncErr
}

func Get(st Type) (*sql.Stmt, error) {
	stmt, ok := prepared[st]
	if !ok {
		return nil, fmt.Errorf("statement %s not found", st.String())
	}
	return stmt, nil
}

func Close() {
	for _, stmt := range prepared {
		err := stmt.Close()
		if err != nil {
			slog.Warn("error closing prepared statement", slog.Any("error", err))
		}
	}
}
