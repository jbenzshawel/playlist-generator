package spotify

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
