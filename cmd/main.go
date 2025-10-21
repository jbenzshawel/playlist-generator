package main

import (
	"log/slog"

	"github.com/jbenzshawel/playlist-generator/internal/config"
)

func main() {
	logger := slog.Default()

	cfg, err := config.Load()
	if err != nil {
		logger.Error("failed to load config", "error", err)
		panic("config not found")
	}

	slog.Info("started!")
}
