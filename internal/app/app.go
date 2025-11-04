package app

import (
	"context"
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
	commands
}

type commands struct {
	Sources   sources.Commands
	Playlists playlists.Commands
}

func NewApplication(ctx context.Context, cfg config.Config) (Application, func()) {
	cfg, err := config.Load()
	if err != nil {
		panic(fmt.Errorf("failed to load config: %w", err))
	}

	db, dbCloser, err := storage.Initialize(ctx, dsn)
	if err != nil {
		panic(fmt.Errorf("failed to initialize database: %w", err))
	}

	srv := &http.Server{
		Addr:    ":3000",
		Handler: nil, // Use http.DefaultServeMux
	}

	// A callback endpoint is required to complete the OAuth authentication code flow
	go func() {
		err := srv.ListenAndServe()
		if err != nil {
			panic(fmt.Errorf("failed to start http server: %w", err))
		}
	}()

	closer := func() {
		dbCloser()

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := srv.Shutdown(shutdownCtx); err != nil {
			slog.Warn("error shutting down server", slog.Any("error", err))
		}
	}

	spotifyClient := setupSpotifyClient(ctx, cfg.SpotifyClient)
	repository := storage.NewRepository(db)

	iprBaseURL, err := url.Parse(cfg.IowaPublicRadio.BaseURL)
	if err != nil {
		panic(fmt.Errorf("failed to parse IowaPublicRadio.BaseURL: %w", err))
	}

	iprClient := studiooneclient.New(studiooneclient.Config{
		BaseURL: iprBaseURL,
	})

	return Application{
		commands: commands{
			Sources:   sources.NewCommands(iprClient, repository),
			Playlists: playlists.NewCommands(spotifyClient, repository),
		},
	}, closer
}

func setupSpotifyClient(ctx context.Context, clientConfig config.OAuthClient) *spotifyclient.Client {
	auth := oauth.NewAuthenticator(oauth.AuthenticatorConfig{
		ClientID:     clientConfig.ClientID,
		ClientSecret: clientConfig.ClientSecret,
		AuthURL:      clientConfig.AuthURL,
		TokenURL:     clientConfig.TokenURL,
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
	completeAuthHandler := auth.GetAuthCodeCallbackHandler(ctx, chOAuthClient)

	http.HandleFunc("/callback", completeAuthHandler)

	fmt.Printf("Click the following URL to complete spotify login: %s\n", loginURL)

	select {
	case <-ctx.Done():
	case spotifyOAuthClient := <-chOAuthClient:
		spotifyClientBaseURL, err := url.Parse(clientConfig.BaseURL)
		if err != nil {
			panic(fmt.Errorf("failed to parse SpotifyClient.BaseURL: %w", err))
		}

		return spotifyclient.New(spotifyclient.Config{
			BaseURL: spotifyClientBaseURL,
			Client:  spotifyOAuthClient,
		})
	}

	return nil
}

type RunConfig struct {
	Mode     Mode
	Date     string
	Month    string
	Interval time.Duration
}

func (a Application) Run(ctx context.Context, cfg RunConfig) {
	switch cfg.Mode {
	case SingleMode:
		if cfg.Month != "" {
			a.genStudioOneSpotifyPlaylistForMonth(ctx, cfg.Month)
		} else {
			a.genStudioOneSpotifyPlaylistsForDay(ctx, cfg.Date)
		}
	case RecurringMode:
		a.startRecurringJob(ctx, cfg.Interval)
	default:
		panic(fmt.Errorf("unknown mode %q", cfg.Mode))
	}

}

func (a Application) startRecurringJob(ctx context.Context, interval time.Duration) {
	slog.Info("starting recurring job", slog.String("interval", fmt.Sprintf("%v minutes", interval.Minutes())))

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	done := make(chan bool)

	go func() {
		for {
			select {
			case <-ticker.C:
				date := time.Now().Format(dateformat.YearMonthDay)
				a.genStudioOneSpotifyPlaylistsForDay(ctx, date)
			case <-ctx.Done():
				slog.Info("stopping recurring job")
				done <- true
				return
			}
		}
	}()

	<-done
}

func (a Application) genStudioOneSpotifyPlaylistForMonth(ctx context.Context, month string) {
	date, err := time.Parse(dateformat.YearMonth, month)
	if err != nil {
		panic(fmt.Errorf("invalid single mode month - YYYY-MM format expected: %w", err))
	}

	end := date.AddDate(0, 1, 0)
	for date.Before(end) {
		select {
		case <-ctx.Done():
		default:
			a.genStudioOneSpotifyPlaylistsForDay(ctx, date.Format(dateformat.YearMonthDay))

			date = date.AddDate(0, 0, 1)
		}
	}
}

func (a Application) genStudioOneSpotifyPlaylistsForDay(ctx context.Context, date string) {
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
