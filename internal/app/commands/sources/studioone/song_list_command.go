package studioone

import (
	"context"
	"log/slog"
	"time"

	"github.com/jbenzshawel/playlist-generator/internal/app/commands/sources/studioone/models"
	"github.com/jbenzshawel/playlist-generator/internal/common/decorator"
	"github.com/jbenzshawel/playlist-generator/internal/domain"
)

const (
	timeLayout = "01-02-2006 15:04:05"
)

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
			queryer:            provider,
			songRepository:     repository.Song(),
			pubRadioRepository: repository.SongSource(),
		},
		repository,
	)
}

type songGetter interface {
	GetSongs(ctx context.Context, date string) (models.Collection, error)
}

type songListCommand struct {
	queryer            songGetter
	songRepository     domain.SongRepository
	pubRadioRepository domain.SongSourceRepository
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
		// TODO: filter on programs
		for _, s := range item.Playlist {
			song, err := domain.NewSong(s.Artist, s.Track, s.Album, s.UPC)
			if err != nil {
				slog.Warn("song skipped", slog.Any("error", err))
				continue
			}

			songs = append(songs, song)

			var programName string
			if item.Program != nil {
				programName = item.Program.Name
			}

			parsedTime, err := time.Parse(timeLayout, s.EndTime)
			if err != nil {
				slog.Warn("song skipped", slog.Any("error", err))
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

	err = d.pubRadioRepository.BulkInsert(ctx, pubRadioSongs)
	if err != nil {
		return nil, err
	}

	return nil, nil
}
