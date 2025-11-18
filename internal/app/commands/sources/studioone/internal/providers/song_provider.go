package providers

import (
	"context"
	"log/slog"
	"strings"
	"time"

	"github.com/jbenzshawel/playlist-generator/internal/app/commands/sources/studioone/models"
	"github.com/jbenzshawel/playlist-generator/internal/common/dateformat"
	"github.com/jbenzshawel/playlist-generator/internal/domain"
)

var supportedPrograms = map[string]struct{}{
	"Blue Avenue":           {},
	"Blues Before Sunrise":  {},
	"Jazz Department":       {},
	"Studio One":            {},
	"Studio One Tracks":     {},
	"Studio One All Access": {},
	"Tiny Desk Radio":       {},
	"UnderCurrents":         {},
	"World Cafe":            {},
}

type SongGetter interface {
	GetSongs(ctx context.Context, date string) (models.Collection, error)
}

func NewSongProvider(getter SongGetter) *songProvider {
	return &songProvider{
		getter: getter,
	}
}

type songProvider struct {
	getter SongGetter
}

func (s *songProvider) ListSongs(ctx context.Context, date string) ([]domain.Song, []domain.SongSource, error) {
	slog.Info("downloading studio one songs", slog.Any("date", date))

	collection, err := s.getter.GetSongs(ctx, date)
	if err != nil {
		return nil, nil, err
	}

	var songs []domain.Song
	var pubRadioSongs []domain.SongSource

	for _, item := range collection.Items {
		var programName string
		if item.Program != nil {
			programName = item.Program.Name
		}
		if _, ok := supportedPrograms[programName]; !ok {
			slog.Debug("unsupported program", slog.String("program", programName))
			continue
		}

		for _, s := range item.Playlist {
			song, err := domain.NewSong(strings.TrimSpace(s.Artist), strings.TrimSpace(s.Track), strings.TrimSpace(s.Album), s.UPC)
			if err != nil {
				slog.Warn("song skipped", slog.Any("error", err))
				continue
			}

			songs = append(songs, song)

			parsedTime, ok := tryParseTime(s.EndTime)
			if !ok {
				slog.Warn("song skipped", slog.Any("invalidEndTime", s.EndTime))
				continue
			}
			pubRadio := domain.NewSongSource(s.ID, song.SongHash(), domain.StudioOneSourceType, programName, date, parsedTime)

			pubRadioSongs = append(pubRadioSongs, pubRadio)
		}
	}

	slog.Info("found songs", slog.Int("count", len(songs)))

	return songs, pubRadioSongs, nil
}

func tryParseTime(t string) (time.Time, bool) {
	parsedTime, err := time.Parse(time.DateTime, t)
	ok := err != nil
	if !ok {
		parsedTime, err = time.Parse(dateformat.MonthDayYearTime, t)
		ok = err != nil
	}
	return parsedTime, ok
}
