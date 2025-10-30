package studioone

import (
	"context"
	"log/slog"
	"strings"
	"time"

	"github.com/jbenzshawel/playlist-generator/internal/app/commands/sources/studioone/models"
	"github.com/jbenzshawel/playlist-generator/internal/common/dateformat"
	"github.com/jbenzshawel/playlist-generator/internal/common/decorator"
	"github.com/jbenzshawel/playlist-generator/internal/domain"
)

var supportedPrograms = map[string]struct{}{
	"Blue Avenue":           {},
	"Studio One":            {},
	"Studio One Tracks":     {},
	"Studio One All Access": {},
	"Tiny Desk Radio":       {},
	"World Cafe":            {},
}

type SongListCommand struct {
	Date string
}

type SongListCommandHandler decorator.CommandHandler[SongListCommand]

func NewSongListCommand(
	provider songGetter,
	repository domain.Repository,
) SongListCommandHandler {
	return decorator.ApplyDBTransactionDecorator(
		&songListCommand{
			queryer:          provider,
			songRepository:   repository.Song(),
			sourceRepository: repository.SongSource(),
		},
		repository,
	)
}

type songGetter interface {
	GetSongs(ctx context.Context, date string) (models.Collection, error)
}

type songListCommand struct {
	queryer          songGetter
	songRepository   domain.SongRepository
	sourceRepository domain.SongSourceRepository
}

func (d *songListCommand) Execute(ctx context.Context, cmd SongListCommand) (any, error) {
	slog.Info("downloading studio one songs", slog.Any("date", cmd.Date))
	collection, err := d.queryer.GetSongs(ctx, cmd.Date)
	if err != nil {
		return nil, err
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
			pubRadio := domain.NewSongSource(s.ID, song.SongHash(), domain.StudioOneSourceType, programName, cmd.Date, parsedTime)

			pubRadioSongs = append(pubRadioSongs, pubRadio)
		}
	}

	slog.Info("found songs", slog.Int("count", len(songs)))

	err = d.songRepository.BulkInsert(ctx, songs)
	if err != nil {
		return nil, err
	}

	err = d.sourceRepository.BulkInsert(ctx, pubRadioSongs)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func tryParseTime(t string) (time.Time, bool) {
	parsedTime, err := time.Parse(dateformat.YearMonthDayTime, t)
	ok := err != nil
	if !ok {
		parsedTime, err = time.Parse(dateformat.MonthDayYearTime, t)
		ok = err != nil
	}
	return parsedTime, ok
}
