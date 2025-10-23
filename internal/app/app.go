package app

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"net/url"

	"github.com/jbenzshawel/playlist-generator/internal/app/sources/studioone"
	"github.com/jbenzshawel/playlist-generator/internal/config"
	"github.com/jbenzshawel/playlist-generator/internal/infrastructure/clients/iprclient"
	"github.com/jbenzshawel/playlist-generator/internal/infrastructure/storage"
)

// TODO: This convention doesn't fit well with this app. Rethink how to bootstrap things?
type Application struct {
	db *sql.DB

	downloaders downloaders
}

type downloaders struct {
	studioOne studioone.Downloader
}

func NewApplication(cfg config.Config, db *sql.DB) Application {
	iprBaseURL, err := url.Parse(cfg.IowaPublicRadio.BaseURL)
	if err != nil {
		panic(fmt.Errorf("failed to parse IowaPublicRadio.BaseURL: %w", err))
	}

	iprClient := iprclient.New(iprclient.Config{
		BaseURL: iprBaseURL,
	})

	songRepo := storage.NewSongSqlRepository(db)
	pubRadioRepo := storage.NewStudioOneSqlRepository(db)

	return Application{
		db: db,
		downloaders: downloaders{
			studioOne: studioone.NewDownloader(iprClient, songRepo, pubRadioRepo),
		},
	}
}

func (a Application) Run(ctx context.Context, date string) {
	err := storage.InitializeSchema(ctx, a.db)
	if err != nil {
		panic(fmt.Errorf("failed to initialize schema: %w", err))
	}

	err = a.downloaders.studioOne.DownloadSongList(ctx, date)
	if err != nil {
		slog.Error("studio one download song list error", slog.Any("error", err))
	}
}
