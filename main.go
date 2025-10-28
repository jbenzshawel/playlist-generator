package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"time"

	_ "modernc.org/sqlite"

	"github.com/jbenzshawel/playlist-generator/internal/app"
	"github.com/jbenzshawel/playlist-generator/internal/app/config"
	"github.com/jbenzshawel/playlist-generator/internal/common/dateformat"
)

func main() {
	defaultDate := time.Now().Format(dateformat.YearMonthDay)
	dateFlag := flag.String("date", defaultDate, "the date to download songs for in YYYY-MM-DD")

	flag.Parse()

	cfg, err := config.Load()
	if err != nil {
		panic(fmt.Errorf("failed to load config: %w", err))
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer stop()

	application, closer := app.NewApplication(cfg)
	defer closer()

	application.Run(ctx, *dateFlag)
}
