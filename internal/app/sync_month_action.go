package app

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/jbenzshawel/playlist-generator/internal/common/dateformat"
	"github.com/jbenzshawel/playlist-generator/internal/domain"
)

func (a Application) syncMonthAction(ctx context.Context, month string) {
	date, err := time.Parse(dateformat.YearMonth, month)
	if err != nil {
		panic(fmt.Errorf("invalid single mode month - YYYY-MM format expected: %w", err))
	}

	end := date.AddDate(0, 1, 0)
	for date.Before(end) {
		select {
		case <-ctx.Done():
		default:
			day := date.Format(time.DateOnly)
			err = a.syncDayAction(ctx, domain.StudioOneSourceType, day)
			if err != nil {
				slog.Error("gen studio one playlist error", slog.Any("error", err), slog.String("date", day))
			}
			date = date.AddDate(0, 0, 1)
		}
	}
}
