package domain

import "strings"

// SourceType represents the type of song source for a playlist.
type SourceType int

const (
	UnknownSourceType   SourceType = 0
	StudioOneSourceType SourceType = 1
	KRUISourceType      SourceType = 2
	KCCKSourceType      SourceType = 3
)

var sourceTypes = map[SourceType]string{
	UnknownSourceType:   "Unknown",
	StudioOneSourceType: "Studio One",
	KRUISourceType:      "KRUI",
	KCCKSourceType:      "KCCK",
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

func AllSourceTypes() []SourceType {
	return []SourceType{
		StudioOneSourceType,
		KRUISourceType,
		KCCKSourceType,
	}
}

func ParseSourceType(s string) SourceType {
	s = strings.ToLower(s)
	for k, v := range sourceTypes {
		v = strings.ToLower(strings.ReplaceAll(v, " ", ""))
		if s == v {
			return k
		}
	}

	return UnknownSourceType
}
