package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"net/url"
	"time"

	_ "modernc.org/sqlite"

	"github.com/jbenzshawel/playlist-generator/internal/app/playlists/spotify"
	"github.com/jbenzshawel/playlist-generator/internal/app/sources/studioone"
	"github.com/jbenzshawel/playlist-generator/internal/config"
	"github.com/jbenzshawel/playlist-generator/internal/domain"
	"github.com/jbenzshawel/playlist-generator/internal/infrastructure/clients/httpclient/oauth"
	"github.com/jbenzshawel/playlist-generator/internal/infrastructure/clients/spotifyclient"
	"github.com/jbenzshawel/playlist-generator/internal/infrastructure/clients/studiooneclient"
	"github.com/jbenzshawel/playlist-generator/internal/infrastructure/storage"
)

const dsn = "file:db/app.db?_busy_timeout=5000&_pragma=journal_mode(WAL)"

func main() {
	defaultDate := time.Now().Format("2006-01-02")
	var dateFlag = flag.String("date", defaultDate, "the date to download songs for in YYYY-MM-DD")

	cfg, err := config.Load()
	if err != nil {
		panic(fmt.Errorf("failed to load config: %w", err))
	}

	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		panic(fmt.Errorf("failed to open database: %w", err))
	}

	defer db.Close()

	ctx := context.Background()

	err = storage.InitializeSchema(ctx, db)
	if err != nil {
		panic(fmt.Errorf("failed to initialize schema: %w", err))
	}

	go func() {
		err := http.ListenAndServe(":3000", nil)
		if err != nil {
			log.Fatal(err)
		}
	}()

	spotifyClient := setupSpotifyClient(cfg)

	repository := storage.NewRepository(db)
	sources := newSources(cfg, repository)

	err = sources.studioOne.Execute(ctx, studioone.SongListCommand{Date: *dateFlag})
	if err != nil {
		slog.Error("studio one download song list error", slog.Any("error", err))
	}

	// TODO: Run in background?
	err = spotify.NewUpdateTracksHandler(spotifyClient, repository).Execute(ctx, spotify.UpdateTracks{})
	if err != nil {
		slog.Error("spotify track update error", slog.Any("error", err))
	}
}

func setupSpotifyClient(cfg config.Config) *spotifyclient.Client {
	// TODO: conditionally get auth code depending on mode
	// May have configuration mode that runs in background downloading song lists
	// and populating them with spotify metadata
	auth := oauth.NewAuthenticator(oauth.AuthenticatorConfig{
		ClientID:     cfg.Clients.SpotifyClient.ClientID,
		ClientSecret: cfg.Clients.SpotifyClient.ClientSecret,
		AuthURL:      cfg.Clients.SpotifyClient.AuthURL,
		TokenURL:     cfg.Clients.SpotifyClient.TokenURL,
		RedirectURL:  "http://127.0.0.1:3000/callback",
		Scopes: []string{
			"playlist-read-private",
			"playlist-modify-private",
		},
	})

	loginURL, err := auth.AuthCodeURL()
	if err != nil {
		panic(fmt.Errorf("auth code url failed: %w", err))
	}

	chOAuthClient := make(chan *http.Client)
	completeAuthHandler := auth.GetAuthCodeCallbackHandler(chOAuthClient)

	http.HandleFunc("/callback", completeAuthHandler)

	fmt.Printf("Click the following URL to complete spotify login: %s\n", loginURL)

	spotifyOAuthClient := <-chOAuthClient

	spotifyClientBaseURL, err := url.Parse(cfg.SpotifyClient.BaseURL)
	if err != nil {
		panic(fmt.Errorf("failed to parse SpotifyClient.BaseURL: %w", err))
	}

	return spotifyclient.New(spotifyclient.Config{
		BaseURL: spotifyClientBaseURL,
		Client:  spotifyOAuthClient,
	})
}

type downloader struct {
	studioOne studioone.SongListCommandHandler
}

func newSources(cfg config.Config, repos domain.Repository) downloader {
	iprBaseURL, err := url.Parse(cfg.IowaPublicRadio.BaseURL)
	if err != nil {
		panic(fmt.Errorf("failed to parse IowaPublicRadio.BaseURL: %w", err))
	}

	iprClient := studiooneclient.New(studiooneclient.Config{
		BaseURL: iprBaseURL,
	})

	return downloader{
		studioOne: studioone.NewSongListCommand(iprClient, repos),
	}
}
