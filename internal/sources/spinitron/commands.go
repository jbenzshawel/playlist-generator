package spinitron

import (
	"context"

	"github.com/jbenzshawel/playlist-generator/internal/domain"
	"github.com/jbenzshawel/playlist-generator/internal/sources/internal"
	"github.com/jbenzshawel/playlist-generator/internal/sources/spinitron/internal/providers"
)

type Client interface {
	providers.SongGetter
}

func NewListSongsCommand(client Client, sourceType domain.SourceType, repository domain.Repository) internal.SongListCommandHandler {
	p := &sourceProvider{
		source:   sourceType,
		provider: providers.NewSongProvider(client),
	}

	return internal.NewSongListCommand(p, repository)
}

type provider interface {
	ListSongs(ctx context.Context, sourceType domain.SourceType) ([]domain.Song, []domain.SongSource, error)
}

type sourceProvider struct {
	source   domain.SourceType
	provider provider
}

func (s *sourceProvider) ListSongs(ctx context.Context, _ string) ([]domain.Song, []domain.SongSource, error) {
	return s.provider.ListSongs(ctx, s.source)
}
