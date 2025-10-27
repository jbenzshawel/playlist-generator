package studioone

type Collection struct {
	Items []Item `json:"onToday"`
}

type Item struct {
	ID        string   `json:"_id"`
	ProgramID string   `json:"program_id"`
	Date      string   `json:"date"`
	StartTime string   `json:"start_time"`
	EndTime   string   `json:"end_time"`
	Program   *Program `json:"program"`
	Playlist  []Song   `json:"playlist"`
}

type Program struct {
	ProgramID string `json:"program_id"`
	Name      string `json:"name"`
	Format    string `json:"program_format"`
}

type Song struct {
	ID      string `json:"_id"`
	Artist  string `json:"artistName"`
	Track   string `json:"trackName"`
	Album   string `json:"collectionName"`
	EndTime string `json:"_end_time"`
	UPC     string `json:"upc"`
}
