package app

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"os/exec"
	"strconv"
	"time"

	"github.com/pterm/pterm"

	"github.com/jbenzshawel/playlist-generator/internal/app/config"
	"github.com/jbenzshawel/playlist-generator/internal/clients/spinitronclient"
	"github.com/jbenzshawel/playlist-generator/internal/clients/spotifyclient"
	"github.com/jbenzshawel/playlist-generator/internal/clients/studiooneclient"
	"github.com/jbenzshawel/playlist-generator/internal/common/output"
	"github.com/jbenzshawel/playlist-generator/internal/domain"
	"github.com/jbenzshawel/playlist-generator/internal/httpclient/oauth"
	"github.com/jbenzshawel/playlist-generator/internal/playlists"
	"github.com/jbenzshawel/playlist-generator/internal/sources"
	"github.com/jbenzshawel/playlist-generator/internal/storage"
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
	Sources   sources.Commands
	Playlists playlists.Commands

	output output.Output
}

func NewApplication(ctx context.Context, outputMode output.Mode) (Application, func()) {
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
		output:    output.New(outputMode),
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

	confirm := pterm.DefaultInteractiveConfirm
	confirm.DefaultText = "Open spotify auth link in default browser?"
	confirm.DefaultValue = true

	result, err := confirm.Show()
	if err != nil {
		slog.Error("failed to prompt confirmation", slog.Any("error", err))
	}

	if !result {
		panic("login with spotify failed")
	}

	cmd := exec.CommandContext(ctx, "open", loginURL)
	if err = cmd.Start(); err != nil {
		panic(fmt.Errorf("failed to open spotify auth link: %w", err))
	}

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
	Action      Action
	Date        string
	Month       string
	Interval    time.Duration
	NumTracks   int
	SongSource  string
	HumanOutput bool
}

func (a Application) Run(ctx context.Context, cfg RunConfig) {
	var sourceTypes []domain.SourceType
	if cfg.SongSource == "" {
		sourceTypes = domain.AllSourceTypes()
	} else {
		sourceType := domain.ParseSourceType(cfg.SongSource)
		if sourceType == domain.UnknownSourceType {
			panic(fmt.Errorf("unknown source type: %s", cfg.SongSource))
		}
		sourceTypes = []domain.SourceType{sourceType}
	}

	switch cfg.Action {
	case SyncDayAction:
		var results []syncDayResult
		for _, st := range sourceTypes {
			res, err := a.syncDayAction(ctx, st, cfg.Date)
			if err != nil {
				slog.Error("sync day error", slog.Any("error", err), slog.String("date", cfg.Date))
			}
			results = append(results, res)
		}
		if len(results) > 1 {
			a.outputSourcesSyncSummary(results)
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

func (a Application) outputSourcesSyncSummary(results []syncDayResult) {
	a.output.Println("\n" + pterm.LightCyan("Sync for all sources complete. Summary of tracks added:"))

	var tableData [][]string
	header := []string{"Playlist name", "Description", "Tracks Added", "Total Tracks"}

	tableData = append(tableData, header)
	for _, r := range results {
		if r.TracksAdded == 0 {
			continue
		}

		tableData = append(tableData, []string{
			r.PlaylistName,
			r.SourceType.Description(),
			strconv.Itoa(r.TracksAdded),
			strconv.Itoa(r.TotalTracks),
		})
	}

	a.output.Table(tableData)
}
