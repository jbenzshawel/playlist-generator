package storage

import (
	"database/sql"
	"github.com/jbenzshawel/playlist-generator/internal/domain"
)

var sourceTypeSchema string = `CREATE TABLE IF NOT EXISTS source_types (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL
);`

func initSourceTypes(db *sql.DB) error {
	for _, st := range domain.AllSourceTypes() {
		query := `INSERT INTO source_types (ID, Name)
		SELECT ?, ?
		WHERE NOT EXISTS (SELECT 1 FROM source_types WHERE id = ?);`

		_, err := db.Exec(query, int(st), st.String(), int(st))
		if err != nil {
			return err
		}
	}

	return nil
}
