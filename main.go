package main

import (
	"context"
	"flag"
	"fmt"
	"time"

	_ "modernc.org/sqlite"

	"github.com/jbenzshawel/playlist-generator/internal/app"
	"github.com/jbenzshawel/playlist-generator/internal/app/config"
	"github.com/jbenzshawel/playlist-generator/internal/common/dateformat"
)

func main() {
	defaultDate := time.Now().Format(dateformat.YearMonthDay)
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
