package domain

type PlaylistDateScope int

const (
	UnknownPlaylistDateScope PlaylistDateScope = 0
	DayPlaylistDateScope     PlaylistDateScope = 1
	MonthPlaylistDateScope   PlaylistDateScope = 2
	YearPlaylistDateScope    PlaylistDateScope = 3
)

var PlaylistDateScopes = map[PlaylistDateScope]string{
	UnknownPlaylistDateScope: "Unknown",
	DayPlaylistDateScope:     "Day",
	MonthPlaylistDateScope:   "Month",
	YearPlaylistDateScope:    "Year",
}

func (t PlaylistDateScope) String() string {
	s, ok := PlaylistDateScopes[t]
	if !ok {
		return "Unknown"
	}
	return s
}

func (t PlaylistDateScope) IsValid() bool {
	_, ok := PlaylistDateScopes[t]
	return ok
}

func AllPlaylistDateScopes() []PlaylistDateScope {
	return []PlaylistDateScope{
		DayPlaylistDateScope,
		MonthPlaylistDateScope,
		YearPlaylistDateScope,
	}
}
