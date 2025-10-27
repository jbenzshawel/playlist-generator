package spotify

import (
	"log/slog"
	"strings"
)

type SearchTrackResponse struct {
	Tracks TrackCollection `json:"tracks"`
}

type TrackCollection struct {
	Total int     `json:"total"`
	Items []Track `json:"items"`
}

type Track struct {
	ID          string      `json:"id"`
	Album       Album       `json:"album"`
	Artists     []Artist    `json:"artists"`
	ExternalIDs ExternalIDs `json:"external_ids"`
	URI         string      `json:"uri"`
	IsPlayable  bool        `json:"is_playable"`
	Name        string      `json:"name"`
}

func (t Track) LogValue() slog.Value {
	var artists []string
	for _, a := range t.Artists {
		artists = append(artists, a.Name)
	}

	return slog.GroupValue(
		slog.String("artist", strings.Join(artists, ", ")),
		slog.String("track", t.Name),
	)
}

type AlbumType string

var (
	AlbumAlbumType       AlbumType = "album"
	SingleAlbumType      AlbumType = "single"
	CompilationAlbumType AlbumType = "compilation"
)

type Album struct {
	ID         string    `json:"id"`
	URI        string    `json:"uri"`
	AlbumType  AlbumType `json:"album_type"`
	IsPlayable bool      `json:"is_playable"`
	Name       string    `json:"name"`
}

type ExternalIDs struct {
	ISRC string `json:"isrc"`
	EAN  string `json:"ean"`
	UPC  string `json:"upc"`
}

type Artist struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	URI  string `json:"uri"`
}

type User struct {
	DisplayName  string            `json:"display_name"`
	ExternalURLs map[string]string `json:"external_urls"`
	Endpoint     string            `json:"href"`
	ID           string            `json:"id"`
	URI          string            `json:"uri"`
}

type PlaylistTracks struct {
	Endpoint string `json:"href"`
	Total    int    `json:"total"`
}

type CreatePlaylistRequest struct {
	Name          string `json:"name"`
	Public        bool   `json:"public"`
	Description   string `json:"description"`
	Collaborative bool   `json:"collaborative"`
}

type SimplePlaylist struct {
	Collaborative bool              `json:"collaborative"`
	Description   string            `json:"description"`
	ExternalURLs  map[string]string `json:"external_urls"`
	Endpoint      string            `json:"href"`
	ID            string            `json:"id"`
	Name          string            `json:"name"`
	Owner         User              `json:"owner"`
	IsPublic      bool              `json:"public"`
	SnapshotID    string            `json:"snapshot_id"`
	Tracks        PlaylistTracks    `json:"tracks"`
	URI           string            `json:"uri"`
}

type AddItemsToPlaylistRequest struct {
	URIs []string `json:"uris"`
}

type PlaylistSnapshot struct {
	SnapshotID string `json:"snapshot_id"`
}
