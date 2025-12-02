package domain

import "strings"

// SourceType represents the type of song source for a playlist.
type SourceType int

const (
	UnknownSourceType SourceType = 0
	// StudioOneSourceType is the IPR music stations
	StudioOneSourceType SourceType = 1
	// KRUISourceType is University of Iowa's college radio station
	KRUISourceType SourceType = 2
	// KCCKSourceType is the Kirkwood Jazz station
	KCCKSourceType SourceType = 3
	// KBEMSourceType is Minnesota's Jazz station Jazz88.FM
	KBEMSourceType SourceType = 4
	// KCSMSourceType is "The Bay Area's Jazz Station"
	KCSMSourceType SourceType = 5
	// EastVillageRadioSourceType is a former pirate radio station broadcasting out of the east village in Manhattan
	EastVillageRadioSourceType SourceType = 6
	// WKCRSourceType is Columbia's college radio station
	WKCRSourceType = 7
	// WDCBSourceType is Chicago's Jazz radio station
	WDCBSourceType = 8
	// KUVOSourceType is Colorado's Jazz radio station
	KUVOSourceType = 9
	// WSUMSourceType is Wisconsin Madison's college radio station
	WSUMSourceType = 10
	// KZSCSourceType is UC Santa Cruz's college radio station
	KZSCSourceType = 11
	// KSPCSourceType is Claremont's college radio station
	KSPCSourceType = 12
)

var sourceTypes = map[SourceType]struct {
	name        string
	description string
}{
	UnknownSourceType:          {name: "Unknown"},
	StudioOneSourceType:        {name: "Studio One", description: "Iowa Public Radio music's station"},
	KRUISourceType:             {name: "KRUI", description: "University of Iowa college radio"},
	KCCKSourceType:             {name: "KCCK", description: "Eastern Iowa's Jazz Station"},
	KBEMSourceType:             {name: "KBEM", description: "Minnesota's Jazz Station"},
	KCSMSourceType:             {name: "KCSM", description: "The Bay Area's Jazz Station"},
	EastVillageRadioSourceType: {name: "EastVillageRadio", description: "East Village Manhattan"},
	WKCRSourceType:             {name: "WKCR", description: "Columbia University college radio"},
	WDCBSourceType:             {name: "WDCB", description: "Chicago's Jazz Station"},
	KUVOSourceType:             {name: "KUVO", description: "Denver's Jazz Station"},
	WSUMSourceType:             {name: "WSUM", description: "University of Wisconsin Madison college radio"},
	KZSCSourceType:             {name: "KZSC", description: "UC Santa Cruz college radio"},
	KSPCSourceType:             {name: "KSPC", description: "Claremont (LA) college radio"},
}

func (t SourceType) String() string {
	s, ok := sourceTypes[t]
	if !ok {
		return "Unknown"
	}
	return s.name
}

func (t SourceType) IsValid() bool {
	_, ok := sourceTypes[t]
	return ok
}

func (t SourceType) Description() string {
	s, ok := sourceTypes[t]
	if !ok {
		return "Unknown"
	}
	return s.description
}

func AllSourceTypes() []SourceType {
	return []SourceType{
		StudioOneSourceType,
		KRUISourceType,
		KCCKSourceType,
		KBEMSourceType,
		KCSMSourceType,
		EastVillageRadioSourceType,
		WKCRSourceType,
		WDCBSourceType,
		KUVOSourceType,
		WSUMSourceType,
		KZSCSourceType,
		KSPCSourceType,
	}
}

func ParseSourceType(s string) SourceType {
	s = strings.ToLower(s)
	for k, v := range sourceTypes {
		name := strings.ToLower(strings.ReplaceAll(v.name, " ", ""))
		if s == name {
			return k
		}
	}

	return UnknownSourceType
}
