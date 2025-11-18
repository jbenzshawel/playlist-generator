package list

import (
	"context"

	"github.com/jbenzshawel/playlist-generator/internal/common/decorator"
	"github.com/jbenzshawel/playlist-generator/internal/domain"
)

type SongListCommand struct {
	Date string
}

type SongListCommandHandler decorator.CommandHandler[SongListCommand]

func NewSongListCommand(
	provider provider,
	repository domain.Repository,
) SongListCommandHandler {
	return decorator.ApplyDBTransactionDecorator(
		&songListCommand{
			provider:         provider,
			songRepository:   repository.Song(),
			sourceRepository: repository.SongSource(),
		},
		repository,
	)
}

type provider interface {
	ListSongs(ctx context.Context, date string) ([]domain.Song, []domain.SongSource, error)
}

type songListCommand struct {
	provider         provider
	songRepository   domain.SongRepository
	sourceRepository domain.SongSourceRepository
}

func (c *songListCommand) Execute(ctx context.Context, cmd SongListCommand) (any, error) {
	songs, pubRadioSongs, err := c.provider.ListSongs(ctx, cmd.Date)
	if err != nil {
		return nil, err
	}

	err = c.songRepository.BulkInsert(ctx, songs)
	if err != nil {
		return nil, err
	}

	err = c.sourceRepository.BulkInsert(ctx, pubRadioSongs)
	if err != nil {
		return nil, err
	}

	return nil, nil
}
