package domain

import "context"

type Repository interface {
	Playlist() PlaylistRepository
	Song() SongRepository
	SongSource() SongSourceRepository
	SpotifyTrack() SpotifyTrackRepository

	Begin(ctx context.Context) error
	Rollback() error
	Commit() error
}
