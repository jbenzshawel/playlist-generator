package storage

import (
	"time"
)

func timeToUTCString(t time.Time) string {
	return t.UTC().Format(time.RFC3339)
}
