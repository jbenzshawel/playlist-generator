package main

import (
	"context"
	"flag"
	"fmt"
	_ "modernc.org/sqlite"
	"time"

	"github.com/jbenzshawel/playlist-generator/internal/app"
	"github.com/jbenzshawel/playlist-generator/internal/app/config"
)

func main() {
	defaultDate := time.Now().Format("2006-01-02")
	var dateFlag = flag.String("date", defaultDate, "the date to download songs for in YYYY-MM-DD")

	cfg, err := config.Load()
	if err != nil {
		panic(fmt.Errorf("failed to load config: %w", err))
	}

	ctx := context.Background()

	application, closer := app.NewApplication(cfg)
	defer closer()

	application.Run(ctx, *dateFlag)
}
