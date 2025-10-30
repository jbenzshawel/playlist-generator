package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
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
	modeFlag := flag.String("mode", string(app.SingleMode), "the mode the generator runs (single or recurring)")
	dateFlag := flag.String("date", defaultDate, "the date to download songs for in YYYY-MM-DD (single mode)")
	monthFlag := flag.String("month", "", "the month to download songs for in YYYY-MM (single mode)")
	intervalFlag := flag.Int("interval", 60, "the interval between downloading songs for in minutes (recurring mode)")
	verboseFlag := flag.Bool("verbose", false, "include detailed logs")

	flag.Parse()

	cfg, err := config.Load()
	if err != nil {
		panic(fmt.Errorf("failed to load config: %w", err))
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer stop()

	application, closer := app.NewApplication(cfg)
	defer closer()

	if *verboseFlag {
		slog.SetLogLoggerLevel(slog.LevelDebug)
	} else {
		slog.SetLogLoggerLevel(slog.LevelInfo)
	}

	application.Run(ctx, app.RunConfig{
		Mode:     app.Mode(*modeFlag),
		Date:     *dateFlag,
		Month:    *monthFlag,
		Interval: time.Duration(*intervalFlag) * time.Minute,
	})
}
