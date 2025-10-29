package app

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"time"

	"github.com/jbenzshawel/playlist-generator/internal/app/commands/playlists"
	"github.com/jbenzshawel/playlist-generator/internal/app/commands/playlists/spotify"
	"github.com/jbenzshawel/playlist-generator/internal/app/commands/sources"
	"github.com/jbenzshawel/playlist-generator/internal/app/commands/sources/studioone"
	"github.com/jbenzshawel/playlist-generator/internal/app/config"
	"github.com/jbenzshawel/playlist-generator/internal/common/dateformat"
	"github.com/jbenzshawel/playlist-generator/internal/infrastructure/clients/httpclient/oauth"
	"github.com/jbenzshawel/playlist-generator/internal/infrastructure/clients/spotifyclient"
	"github.com/jbenzshawel/playlist-generator/internal/infrastructure/clients/studiooneclient"
	"github.com/jbenzshawel/playlist-generator/internal/infrastructure/storage"
)

const dsn = "file:db/app.db?_busy_timeout=5000&_pragma=journal_mode(WAL)"

type Mode string

const (
	SingleMode    Mode = "single"
	RecurringMode Mode = "recurring"
)

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

	// A callback endpoint is required to complete the OAuth authentication code flow
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

func setupSpotifyClient(cfg config.Config) *spotifyclient.Client {
	// TODO: conditionally get auth code depending on Mode
	// May have configuration Mode that runs in background downloading song lists
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

type RunConfig struct {
	Mode     Mode
	Date     string
	Interval time.Duration
}

func (a Application) Run(ctx context.Context, cfg RunConfig) {
	err := storage.InitializeSchema(ctx, a.db)
	if err != nil {
		panic(fmt.Errorf("failed to initialize schema: %w", err))
	}

	switch cfg.Mode {
	case SingleMode:
		a.genStudioOneSpotifyPlaylists(ctx, cfg.Date)
	case RecurringMode:
		a.startRecurringJob(ctx, cfg.Interval)
	default:
		panic(fmt.Errorf("unknown mode %q", cfg.Mode))
	}

}

func (a Application) genStudioOneSpotifyPlaylists(ctx context.Context, date string) {
	slog.Info("adding songs from Studio One to Spotify playlist", slog.String("date", date))

	_, err := a.Sources.StudioOne.ListSongs.Execute(ctx, studioone.SongListCommand{Date: date})
	if err != nil {
		slog.Error("studio one download song list error", slog.Any("error", err))
	}

	_, err = a.Playlists.Spotify.SearchTracks.Execute(ctx, spotify.SearchTracksCommand{})
	if err != nil {
		slog.Error("spotify track update error", slog.Any("error", err))
	}

	createRes, err := a.Playlists.Spotify.CreatePlaylist.Execute(ctx, spotify.CreatePlaylistCommand{
		Date: date,
	})
	if err != nil {
		slog.Error("create spotify playlist error", slog.Any("error", err))
	}

	_, err = a.Playlists.Spotify.SyncPlaylist.Execute(ctx, spotify.SyncPlaylistCommand{
		Playlist: createRes.Playlist,
		Date:     date,
	})
	if err != nil {
		slog.Error("sync spotify playlist error", slog.Any("error", err))
	}
}

func (a Application) startRecurringJob(ctx context.Context, interval time.Duration) {
	slog.Info("starting recurring job", slog.String("interval", fmt.Sprintf("%v minutes", interval.Minutes())))

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	done := make(chan bool)

	go func() {
		for {
			for {
				select {
				case <-ticker.C:
					date := time.Now().Format(dateformat.YearMonthDay)
					a.genStudioOneSpotifyPlaylists(ctx, date)
				case <-ctx.Done():
					slog.Info("stopping recurring job")
					done <- true
					return
				}
			}
		}
	}()

	<-done
}
