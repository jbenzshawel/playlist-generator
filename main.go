package main

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"os/signal"
	"time"

	"github.com/pterm/pterm"

	"github.com/jbenzshawel/playlist-generator/internal/app"
	"github.com/jbenzshawel/playlist-generator/internal/common/output"
)

func main() {
	defaultDate := time.Now().Format(time.DateOnly)

	actionFlag := flag.String("action", string(app.SyncDayAction), "the action the generator runs (syncDay, syncMonth, recurring, or random)")
	dateFlag := flag.String("date", defaultDate, "the date to download songs for in YYYY-MM-DD (syncDay action)")
	monthFlag := flag.String("month", "", "the month to download songs for in YYYY-MM (syncMonth action)")
	intervalFlag := flag.Int("interval", 60, "the interval between downloading songs for in minutes (recurring action)")
	numTracksFlag := flag.Int("numTracks", 50, "the number of random tracks to include in the random tracks playlist (random action)")
	songSourceFlag := flag.String("source", "", "the source type to download songs from. If none specified all sources will be synced")
	verboseFlag := flag.Bool("verbose", false, "include detailed logs (human readable format when false)")

	flag.Parse()

	var outputMode output.Mode
	if *verboseFlag {
		outputMode = output.MachineMode
		slog.SetLogLoggerLevel(slog.LevelDebug)
	} else {
		outputMode = output.HumanMode
		logger := &pterm.DefaultLogger
		logger.Level = pterm.LogLevelError

		handler := pterm.NewSlogHandler(&pterm.DefaultLogger)

		slog.SetDefault(slog.New(handler))
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer stop()

	application, closer := app.NewApplication(ctx, outputMode)
	defer closer()

	select {
	case <-ctx.Done():
	default:
		application.Run(ctx, app.RunConfig{
			Action:     app.Action(*actionFlag),
			Date:       *dateFlag,
			Month:      *monthFlag,
			Interval:   time.Duration(*intervalFlag) * time.Minute,
			NumTracks:  *numTracksFlag,
			SongSource: *songSourceFlag,
		})
	}
}
