package providers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"regexp"
	"strings"
	"time"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"

	"github.com/jbenzshawel/playlist-generator/internal/common/dateformat"
	"github.com/jbenzshawel/playlist-generator/internal/domain"
)

var today = time.Now().Format(time.DateOnly)

var (
	regexSanitize = regexp.MustCompile(`[\n\t\r]+`)

	errSongParseFailed     = errors.New("song parse failed")
	errPlayTimeParseFailed = errors.New("playtime parse failed")
)

type SongGetter interface {
	ScrapePlaylist(ctx context.Context, source string) ([]byte, error)
}

func NewSongProvider(getter SongGetter) *songProvider {
	return &songProvider{
		getter: getter,
	}
}

type songProvider struct {
	getter SongGetter
}

func (s *songProvider) ListSongs(ctx context.Context, sourceType domain.SourceType) ([]domain.Song, []domain.SongSource, error) {
	result, err := s.scrapePlaylist(ctx, fmt.Sprintf("/%s", sourceType), sourceType)
	if err != nil {
		return nil, nil, err
	}

	for _, prev := range result.previous {
		prevRes, err := s.scrapePlaylist(ctx, prev, sourceType)
		if err != nil {
			return nil, nil, err
		}
		result.songs = append(result.songs, prevRes.songs...)
		result.sources = append(result.sources, prevRes.sources...)
	}

	return result.songs, result.sources, nil
}

func (s *songProvider) scrapePlaylist(ctx context.Context, source string, sourceType domain.SourceType) (parseResult, error) {
	raw, err := s.getter.ScrapePlaylist(ctx, source)
	if err != nil {
		return parseResult{}, err
	}

	result, err := parseSongData(raw, sourceType)
	if err != nil {
		return parseResult{}, err
	}
	return result, nil
}

type parseResult struct {
	songs    []domain.Song
	sources  []domain.SongSource
	previous []string
}

func parseSongData(raw []byte, sourceType domain.SourceType) (parseResult, error) {
	if len(raw) == 0 {
		return parseResult{}, nil
	}

	doc, err := html.Parse(bytes.NewReader(raw))
	if err != nil {
		return parseResult{}, err
	}

	var programName string
	result := parseResult{}
	for n := range doc.Descendants() {
		if isProgramNode(n) {
			programName = parseProgram(n)
		}

		if isSongRow(n) {
			song, sourceSongID, err := parseSong(n)
			if err != nil {
				slog.Warn("failed to parse song", slog.Any("error", err))
				continue
			}

			timePlayed, err := parseTimePlayed(n)
			if err != nil {
				slog.Warn("failed to parse time played", slog.Any("error", err))
				continue
			}

			result.songs = append(result.songs, song)
			result.sources = append(result.sources, domain.NewSongSource(
				sourceSongID,
				song.SongHash(),
				sourceType,
				programName,
				today,
				timePlayed,
			))
		}

		if isRecentPlaylist(n) {
			result.previous = parseRecentPlaylists(n)
			continue
		}
	}

	return result, nil
}

func isProgramNode(n *html.Node) bool {
	classNames := nodeClassNames(n)
	_, hasShowTitle := classNames["show-title"]

	return n.Type == html.ElementNode &&
		n.DataAtom == atom.H3 &&
		hasShowTitle
}

func parseProgram(n *html.Node) string {
	if n.FirstChild != nil && n.FirstChild.NextSibling != nil {
		a := n.FirstChild.NextSibling
		if a.DataAtom == atom.A {
			return sanitize(a.FirstChild.Data)
		}
	}
	return ""
}

func isSongRow(n *html.Node) bool {
	classNames := nodeClassNames(n)
	_, hasSpinItem := classNames["spin-item"]

	return n.Type == html.ElementNode &&
		n.DataAtom == atom.Tr &&
		hasSpinItem
}

type spin struct {
	Arist *string `json:"a"`
	Song  *string `json:"s"`
	Album *string `json:"r"`
}

func parseSong(n *html.Node) (domain.Song, string, error) {
	var dataSpin, dataKey string
	for _, attr := range n.Attr {
		switch attr.Key {
		case "data-spin":
			dataSpin = html.UnescapeString(attr.Val)
		case "data-key":
			dataKey = attr.Val
		}
	}

	var s spin
	err := json.Unmarshal([]byte(dataSpin), &s)
	if err != nil {
		return domain.Song{}, "", fmt.Errorf("%w: %w", errSongParseFailed, err)
	}

	if s.Arist == nil && s.Song == nil {
		return domain.Song{}, "", errSongParseFailed
	}

	var album string
	if s.Album != nil {
		album = *s.Album
	}

	song, err := domain.NewSong(*s.Arist, *s.Song, album, "")
	if err != nil {
		return domain.Song{}, "", fmt.Errorf("%w: %w", errSongParseFailed, err)
	}

	return song, dataKey, nil
}

func parseTimePlayed(n *html.Node) (time.Time, error) {
	for desc := range n.ChildNodes() {
		classNames := nodeClassNames(desc)
		if _, ok := classNames["spin-time"]; !ok || desc.FirstChild == nil || desc.FirstChild.FirstChild == nil {
			continue
		}

		playTime := strings.ReplaceAll(sanitize(desc.FirstChild.FirstChild.Data), " ", "")

		parsedTime, err := time.Parse(dateformat.YearMonthDayKitchen, fmt.Sprintf("%s %s", today, playTime))
		if err != nil {
			return time.Time{}, fmt.Errorf("%w: %w", errPlayTimeParseFailed, err)
		}

		return parsedTime, nil
	}

	return time.Time{}, errSongParseFailed
}

func isRecentPlaylist(n *html.Node) bool {
	classNames := nodeClassNames(n)
	_, ok := classNames["recent-playlists"]
	return ok
}

func parseRecentPlaylists(n *html.Node) []string {
	tableNode := findAtomNode(n, atom.Tbody)
	if tableNode == nil {
		return nil
	}

	var tableRows []*html.Node
	tr := tableNode.FirstChild
	for tr != nil {
		if tr.DataAtom == atom.Tr {
			tableRows = append(tableRows, tr)
		}
		tr = tr.NextSibling
	}

	results := make([]string, 0, len(tableRows))
	for _, row := range tableRows {
		// next link is in third column
		var cols []*html.Node
		td := row.FirstChild
		for td != nil {
			if td.DataAtom == atom.Td {
				cols = append(cols, td)
			}
			td = td.NextSibling
		}
		if len(cols) != 3 {
			return nil
		}

		a := findAtomNode(cols[2], atom.A)
		if a == nil {
			return nil
		}

		for _, attr := range a.Attr {
			if attr.Key == "href" {
				results = append(results, attr.Val)
			}
		}
	}

	return results
}

func sanitize(s string) string {
	return regexSanitize.ReplaceAllString(s, "")
}

func nodeClassNames(n *html.Node) map[string]struct{} {
	var classNames []string
	for _, a := range n.Attr {
		if a.Key == "class" {
			classNames = strings.Split(a.Val, " ")
			break
		}
	}

	classNameLookup := make(map[string]struct{}, len(classNames))
	for _, c := range classNames {
		classNameLookup[c] = struct{}{}
	}

	return classNameLookup
}

func findAtomNode(n *html.Node, a atom.Atom) *html.Node {
	for n != nil {
		if n.DataAtom == a {
			return n
		}

		if n.FirstChild != nil {
			n = n.FirstChild
			continue
		}

		if n.NextSibling != nil {
			n = n.NextSibling
			continue
		}

		// move on to the parent's next node
		n = n.Parent.NextSibling
	}

	return nil
}
