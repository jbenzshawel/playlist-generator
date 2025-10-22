package studioone

import (
	"context"
	"log/slog"
	"time"

	"github.com/jbenzshawel/playlist-generator/internal/domain"
)

const (
	timeLayout = "01-02-2006 15:04:05"
)

type Downloader interface {
	DownloadSongList(ctx context.Context, date string) error
}

type provider interface {
	GetSongs(ctx context.Context, date string) (Collection, error)
}

type downloader struct {
	provider           provider
	songRepository     domain.SongRepository
	pubRadioRepository domain.PublicRadioRepository
}

func NewDownloader(
	provider provider,
	songRepository domain.SongRepository,
	pubRadioRepository domain.PublicRadioRepository,
) *downloader {
	return &downloader{
		provider:           provider,
		songRepository:     songRepository,
		pubRadioRepository: pubRadioRepository,
	}
}

func (d *downloader) DownloadSongList(ctx context.Context, date string) error {
	slog.Info("downloading studio one songs", slog.Any("date", date))
	collection, err := d.provider.GetSongs(ctx, date)
	if err != nil {
		return err
	}

	var songs []domain.Song
	var pubRadioSongs []domain.PublicRadioSong

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
			pubRadio := domain.NewPublicRadioSong(s.ID, song.SongHash(), programName, date, parsedTime)

			pubRadioSongs = append(pubRadioSongs, pubRadio)
		}
	}

	slog.Info("found songs", slog.Int("count", len(songs)))

	err = d.songRepository.BulkInsert(ctx, songs)
	if err != nil {
		return err
	}

	err = d.pubRadioRepository.BulkInsert(ctx, pubRadioSongs)
	if err != nil {
		return err
	}

	// TODO: switch to https://github.com/mattn/go-sqlite3
	// to fix concurrency?
	//g, ctx := errgroup.WithContext(ctx)
	//g.Go(func() error {
	//	return d.songRepository.BulkInsert(ctx, songs)
	//})
	//g.Go(func() error {
	//	return d.pubRadioRepository.BulkInsert(ctx, pubRadioSongs)
	//})
	//
	//err = g.Wait()
	//if err != nil {
	//	return err
	//}

	return nil
}
