package studioone

import (
	"context"
	"log/slog"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/jbenzshawel/playlist-generator/internal/domain"
)

const (
	timeLayout = "01-02-2006 15:04:05"
)

type Provider interface {
	GetSongs(ctx context.Context, date string) (Collection, error)
}

type downloader struct {
	provider               Provider
	songRepository         domain.SongRepository
	pubRadioSongRepository domain.PublicSongRepository
}

func NewDownloader(provider Provider, songRepository domain.SongRepository) *downloader {
	return &downloader{
		provider:       provider,
		songRepository: songRepository,
	}
}

func (d *downloader) DownloadSongList(ctx context.Context, date string) error {
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

	g, ctx := errgroup.WithContext(ctx)
	g.Go(func() error {
		return d.songRepository.BulkInsert(songs)
	})
	g.Go(func() error {
		return d.pubRadioSongRepository.BulkInsert(pubRadioSongs)
	})

	err = g.Wait()
	if err != nil {
		return err
	}

	return nil
}
