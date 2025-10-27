package app

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"

	"github.com/jbenzshawel/playlist-generator/internal/app/commands/playlists"
	"github.com/jbenzshawel/playlist-generator/internal/app/commands/playlists/spotify"
	"github.com/jbenzshawel/playlist-generator/internal/app/commands/sources"
	"github.com/jbenzshawel/playlist-generator/internal/app/commands/sources/studioone"
	"github.com/jbenzshawel/playlist-generator/internal/app/config"
	"github.com/jbenzshawel/playlist-generator/internal/infrastructure/clients/httpclient/oauth"
	"github.com/jbenzshawel/playlist-generator/internal/infrastructure/clients/spotifyclient"
	"github.com/jbenzshawel/playlist-generator/internal/infrastructure/clients/studiooneclient"
	"github.com/jbenzshawel/playlist-generator/internal/infrastructure/storage"
)

const dsn = "file:db/app.db?_busy_timeout=5000&_pragma=journal_mode(WAL)"

type Application struct {
	db *sql.DB

	commands
}

type commands struct {
	Sources   sources.Commands
	Playlists playlists.Commands
}

func NewApplication(cfg config.Config) (Application, func()) {
	cfg, err := config.Load()
	if err != nil {
		panic(fmt.Errorf("failed to load config: %w", err))
	}

	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		panic(fmt.Errorf("failed to open database: %w", err))
	}

	closer := func() {
		err := db.Close()
		if err != nil {
			slog.Warn("error closing database", slog.Any("error", err))
		}
	}

	go func() {
		err := http.ListenAndServe(":3000", nil)
		if err != nil {
			panic(fmt.Errorf("failed to start http server: %w", err))
		}
	}()

	spotifyClient := setupSpotifyClient(cfg)
	repository := storage.NewRepository(db)

	iprBaseURL, err := url.Parse(cfg.IowaPublicRadio.BaseURL)
	if err != nil {
		panic(fmt.Errorf("failed to parse IowaPublicRadio.BaseURL: %w", err))
	}

	iprClient := studiooneclient.New(studiooneclient.Config{
		BaseURL: iprBaseURL,
	})

	return Application{
		db: db,
		commands: commands{
			Sources:   sources.NewCommands(iprClient, repository),
			Playlists: playlists.NewCommands(spotifyClient, repository),
		},
	}, closer
}

func (a Application) Run(ctx context.Context, date string) {
	err := storage.InitializeSchema(ctx, a.db)
	if err != nil {
		panic(fmt.Errorf("failed to initialize schema: %w", err))
	}

	err = a.Sources.StudioOne.ListSongs.Execute(ctx, studioone.SongListCommand{Date: date})
	if err != nil {
		slog.Error("studio one download song list error", slog.Any("error", err))
	}

	err = a.Playlists.Spotify.UpdateTracks.Execute(ctx, spotify.UpdateTracksCommand{})
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
