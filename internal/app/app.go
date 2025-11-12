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

type Action string

const (
	SyncDayAction   Action = "syncDay"
	SyncMonthAction Action = "syncMonth"
	RecurringAction Action = "recurring"
	RandomAction    Action = "random"
)

type Application struct {
	commands
}

type commands struct {
	Sources   sources.Commands
	Playlists playlists.Commands
}

func NewApplication(ctx context.Context) (Application, func()) {
	cfg, err := config.Load()
	if err != nil {
		panic(fmt.Errorf("failed to load config: %w", err))
	}

	store, err := storage.Initialize(ctx, dsn)
	if err != nil {
		panic(fmt.Errorf("failed to initialize database: %w", err))
	}

	// A callback endpoint is required to complete the OAuth authentication code flow
	go func() {
		err := http.ListenAndServe(":3000", nil)
		if err != nil {
			panic(fmt.Errorf("failed to start http server: %w", err))
		}
	}()

	closer := func() {
		store.Close()
	}

	spotifyClient := setupSpotifyClient(ctx, cfg.SpotifyClient)
	repository := storage.NewRepository(store)

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
			"playlist-modify-public",
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
	Action    Action
	Date      string
	Month     string
	Interval  time.Duration
	NumTracks int
}

func (a Application) Run(ctx context.Context, cfg RunConfig) {
	switch cfg.Action {
	case SyncDayAction:
		err := a.genStudioOneSpotifyPlaylistsForDay(ctx, cfg.Date)
		if err != nil {
			slog.Error("gen studio one playlist error", slog.Any("error", err), slog.String("date", cfg.Date))
		}
	case SyncMonthAction:
		a.genStudioOneSpotifyPlaylistForMonth(ctx, cfg.Month)
	case RecurringAction:
		a.startRecurringJob(ctx, cfg.Interval)
	case RandomAction:
		err := a.randomPlaylist(ctx, cfg.NumTracks)
		if err != nil {
			slog.Error("update random playlist error", slog.Any("error", err))
		}
	default:
		panic(fmt.Errorf("unknown action %q", cfg.Action))
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
				date := time.Now().Format(time.DateOnly)
				err := a.genStudioOneSpotifyPlaylistsForDay(ctx, date)
				if err != nil {
					slog.Error("gen studio one playlist error", slog.Any("error", err), slog.String("date", date))
				}
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
			day := date.Format(time.DateOnly)
			err = a.genStudioOneSpotifyPlaylistsForDay(ctx, day)
			if err != nil {
				slog.Error("gen studio one playlist error", slog.Any("error", err), slog.String("date", day))
			}
			date = date.AddDate(0, 0, 1)
		}
	}
}

func (a Application) genStudioOneSpotifyPlaylistsForDay(ctx context.Context, date string) error {
	slog.Info("adding songs from Studio One to Spotify playlist", slog.String("date", date))

	_, err := a.Sources.StudioOne.ListSongs.Execute(ctx, studioone.SongListCommand{Date: date})
	if err != nil {
		return fmt.Errorf("studio one download song list error: %w", err)
	}

	_, err = a.Playlists.Spotify.SearchTracks.Execute(ctx, spotify.SearchTracksCommand{})
	if err != nil {
		return fmt.Errorf("spotify track update error: %w", err)
	}

	createRes, err := a.Playlists.Spotify.CreatePlaylist.Execute(ctx, spotify.CreatePlaylistCommand{
		Date: date,
	})
	if err != nil {
		return fmt.Errorf("create spotify playlist error: %w", err)
	}

	_, err = a.Playlists.Spotify.SyncPlaylist.Execute(ctx, spotify.SyncPlaylistCommand{
		Playlist: createRes.Playlist,
		Date:     date,
	})
	if err != nil {
		return fmt.Errorf("sync spotify playlist error: %w", err)
	}

	return err
}

func (a Application) randomPlaylist(ctx context.Context, numTracks int) error {
	slog.Info("updating random playlist with new random tracks", slog.Int("numTracks", numTracks))

	_, err := a.Playlists.Spotify.RandomTracksPlaylist.Execute(ctx, spotify.RandomTracksPlaylistCommand{
		NumTracks: numTracks,
	})
	if err != nil {
		return err
	}

	return nil
}
