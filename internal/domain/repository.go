package domain

import "context"

type Repository interface {
	Songs() SongRepository
	SongSource() SongSourceRepository
	SpotifyTracks() SpotifyTrackRepository

	Begin(ctx context.Context) error
	Rollback() error
	Commit() error
}
