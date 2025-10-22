package main

import (
	"github.com/jbenzshawel/playlist-generator/internal/app/config"
	"log/slog"
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
