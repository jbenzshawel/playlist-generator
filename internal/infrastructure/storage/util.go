package storage

import (
	"log/slog"
	"time"
)

func timeToUTCString(t time.Time) string {
	return t.UTC().Format(time.RFC3339)
}

func utcStringToTime(s string) (time.Time, error) {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return time.Time{}, err
	}

	return t.Local(), nil
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

type closer interface {
	Close() error
}

func closeStatement(c closer) {
	err := c.Close()
	if err != nil {
		slog.Warn("error closing prepared statement", slog.Any("error", err))
	}
}
