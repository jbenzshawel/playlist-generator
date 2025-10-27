package storage

import (
	"database/sql"

	"github.com/jbenzshawel/playlist-generator/internal/domain"
)

var playlistTypeSchema string = `CREATE TABLE IF NOT EXISTS playlist_types (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL
);`

func initPlaylistTypes(db *sql.DB) error {
	for _, pt := range domain.AllSourceTypes() {
		query := `INSERT INTO playlist_types (ID, Name)
		SELECT ?, ?
		WHERE NOT EXISTS (SELECT 1 FROM playlist_types WHERE id = ?);`

		_, err := db.Exec(query, int(pt), pt.String(), int(pt))
		if err != nil {
			return err
		}
	}

	return nil
}
