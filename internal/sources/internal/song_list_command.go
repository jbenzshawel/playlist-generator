package internal

import (
	"context"

	"github.com/jbenzshawel/playlist-generator/internal/common/decorator"
	"github.com/jbenzshawel/playlist-generator/internal/domain"
)

type SongListCommand struct {
	Date string
}

type SongListCommandResult struct {
	FoundCount int
}

type SongListCommandHandler decorator.CommandWithResultHandler[SongListCommand, SongListCommandResult]

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

func (c *songListCommand) Execute(ctx context.Context, cmd SongListCommand) (SongListCommandResult, error) {
	songs, sources, err := c.provider.ListSongs(ctx, cmd.Date)
	if err != nil {
		return SongListCommandResult{}, err
	}

	err = c.songRepository.BulkInsert(ctx, songs)
	if err != nil {
		return SongListCommandResult{}, err
	}

	err = c.sourceRepository.BulkInsert(ctx, sources)
	if err != nil {
		return SongListCommandResult{}, err
	}

	return SongListCommandResult{FoundCount: len(songs)}, nil
}
