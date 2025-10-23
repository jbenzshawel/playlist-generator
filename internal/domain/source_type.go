package domain

// SourceType represents the type of song source for a playlist.
type SourceType int

const (
	UnknownSourceType   SourceType = 0
	StudioOneSourceType SourceType = 1
)

var sourceTypes = map[SourceType]string{
	UnknownSourceType:   "Unknown",
	StudioOneSourceType: "Studio One",
}

func (t SourceType) String() string {
	s, ok := sourceTypes[t]
	if !ok {
		return "Unknown"
	}
	return s
}

func (t SourceType) IsValid() bool {
	_, ok := sourceTypes[t]
	return ok
}
