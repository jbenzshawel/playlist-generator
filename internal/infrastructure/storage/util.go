package storage

import (
	"time"
)

func timeToUTCString(t time.Time) string {
	return t.UTC().Format(time.RFC3339)
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
