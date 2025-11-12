package services

import (
	"github.com/jbenzshawel/playlist-generator/internal/app/commands/playlists/spotify/internal/services/internal/mutators"
	"github.com/jbenzshawel/playlist-generator/internal/app/commands/playlists/spotify/internal/services/internal/providers"
)

type Client interface {
	providers.TrackSearcher
	providers.TrackGetter
	mutators.PlaylistCreator
	mutators.TrackAdderRemover
}

type PlaylistService interface {
	providers.PlaylistTrackProvider
	mutators.PlaylistTrackMutator
	mutators.CreatePlaylistMutator
}

type playlistService struct {
	providers.PlaylistTrackProvider
	mutators.PlaylistTrackMutator
	mutators.CreatePlaylistMutator
}

func NewPlaylistService(client Client) PlaylistService {
	return &playlistService{
		PlaylistTrackProvider: providers.NewPlaylistTrackProvider(client),
		PlaylistTrackMutator:  mutators.NewPlaylistTrackMutator(client),
		CreatePlaylistMutator: mutators.NewCreatePlaylistMutator(client),
	}
}

type SearchService interface {
	providers.SearchTrackProvider
}

type searchService struct {
	providers.SearchTrackProvider
}

func NewSearchService(client Client) SearchService {
	return &searchService{
		SearchTrackProvider: providers.NewSearchTrackProvider(client),
	}
}
