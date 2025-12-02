package app

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/jbenzshawel/playlist-generator/internal/domain"
)

func (a Application) recurringAction(ctx context.Context, interval time.Duration) {
	slog.Info("starting recurring job", slog.String("interval", fmt.Sprintf("%v minutes", interval.Minutes())))

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	done := make(chan bool)

	go func() {
		for {
			select {
			case <-ticker.C:
				for _, sourceType := range domain.AllSourceTypes() {
					date := time.Now().Format(time.DateOnly)
					err := a.syncDayAction(ctx, sourceType, date)
					if err != nil {
						slog.Error("gen playlist error",
							slog.String("sourceType", sourceType.String()),
							slog.Any("error", err),
							slog.String("date", date),
						)
					}
				}

			case <-ctx.Done():
				slog.Info("stopping recurringAction job")
				done <- true
				return
			}
		}
	}()

	<-done
}
