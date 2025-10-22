package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"time"

	_ "modernc.org/sqlite" // Registers the "sqlite" driver

	"github.com/jbenzshawel/playlist-generator/internal/app"
	"github.com/jbenzshawel/playlist-generator/internal/config"
)

const dsn = "file:db/app.db?_busy_timeout=5000&_pragma=journal_mode(WAL)"

func main() {
	defaultDate := time.Now().Format("2006-01-02")
	var dateFlag = flag.String("date", defaultDate, "the date to download songs for in YYYY-MM-DD")

	cfg, err := config.Load()
	if err != nil {
		panic(fmt.Errorf("failed to load config: %w", err))
	}

	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		panic(fmt.Errorf("failed to open database: %w", err))
	}

	defer db.Close()

	ctx := context.Background()

	app.NewApplication(cfg, db).
		Run(ctx, *dateFlag)
}
