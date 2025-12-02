package app

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"time"

	"github.com/jbenzshawel/playlist-generator/internal/app/config"
	"github.com/jbenzshawel/playlist-generator/internal/clients/spinitronclient"
	"github.com/jbenzshawel/playlist-generator/internal/clients/spotifyclient"
	"github.com/jbenzshawel/playlist-generator/internal/clients/studiooneclient"
	"github.com/jbenzshawel/playlist-generator/internal/domain"
	"github.com/jbenzshawel/playlist-generator/internal/httpclient/oauth"
	"github.com/jbenzshawel/playlist-generator/internal/playlists"
	"github.com/jbenzshawel/playlist-generator/internal/sources"
	"github.com/jbenzshawel/playlist-generator/internal/storage"
)

const dsn = "file:db/app.db?_busy_timeout=5000&_pragma=journal_mode(WAL)"

type Action string

const (
	SyncDayAction   Action = "syncDayAction"
	SyncMonthAction Action = "syncMonthAction"
	RecurringAction Action = "recurringAction"
	RandomAction    Action = "randomAction"
)

type Application struct {
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

	spinBaseURL, err := url.Parse(cfg.Spinitron.BaseURL)
	if err != nil {
		panic(fmt.Errorf("failed to parse Spinitron.BaseURL: %w", err))
	}

	spinClient := spinitronclient.New(spinitronclient.Config{
		BaseURL: spinBaseURL,
	})

	return Application{
		Sources:   sources.NewCommands(iprClient, spinClient, repository),
		Playlists: playlists.NewCommands(spotifyClient, repository),
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
	Action     Action
	Date       string
	Month      string
	Interval   time.Duration
	NumTracks  int
	SongSource string
}

func (a Application) Run(ctx context.Context, cfg RunConfig) {
	sourceType := domain.ParseSourceType(cfg.SongSource)
	if sourceType == domain.UnknownSourceType {
		panic(fmt.Errorf("unknown source type: %s", cfg.SongSource))
	}

	sourceTypes := []domain.SourceType{sourceType}
	if cfg.SongSource == "" {
		sourceTypes = domain.AllSourceTypes()
	}

	switch cfg.Action {
	case SyncDayAction:
		for _, st := range sourceTypes {
			err := a.syncDayAction(ctx, st, cfg.Date)
			if err != nil {
				slog.Error("sync day error", slog.Any("error", err), slog.String("date", cfg.Date))
			}
		}
	case SyncMonthAction:
		a.syncMonthAction(ctx, cfg.Month)
	case RecurringAction:
		a.recurringAction(ctx, cfg.Interval)
	case RandomAction:
		err := a.randomAction(ctx, cfg.NumTracks)
		if err != nil {
			slog.Error("update randomAction playlist error", slog.Any("error", err))
		}
	default:
		panic(fmt.Errorf("unknown action %q", cfg.Action))
	}
}
