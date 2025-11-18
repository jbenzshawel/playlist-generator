package providers

import (
	_ "embed"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/jbenzshawel/playlist-generator/internal/common/dateformat"
	"github.com/jbenzshawel/playlist-generator/internal/domain"
)

var (
	//go:embed test_song_list.html
	testSongData []byte

	//go:embed test_song_list_prev.html
	testPrevSongData []byte
)

func TestSongListProvider_ListSongs(t *testing.T) {
	mockGetter := NewMockSongGetter(t)
	mockGetter.EXPECT().ScrapePlaylist(mock.Anything, "/KRUI").Return(testSongData, nil)
	mockGetter.EXPECT().ScrapePlaylist(mock.Anything, "/KRUI/pl/21532308/89-7FM-KRUI-Iowa-City-11-15-25-8-00-AM").Return(testPrevSongData, nil)
	mockGetter.EXPECT().ScrapePlaylist(mock.Anything, "/KRUI/pl/21532084/89-7FM-KRUI-Iowa-City-11-15-25-7-00-AM").Return(nil, nil)
	mockGetter.EXPECT().ScrapePlaylist(mock.Anything, "/KRUI/pl/21531905/89-7FM-KRUI-Iowa-City-11-15-25-6-00-AM").Return(nil, nil)
	mockGetter.EXPECT().ScrapePlaylist(mock.Anything, "/KRUI/pl/21531822/89-7FM-KRUI-Iowa-City-11-15-25-5-03-AM").Return(nil, nil)
	mockGetter.EXPECT().ScrapePlaylist(mock.Anything, "/KRUI/pl/21531600/89-7FM-KRUI-Iowa-City-11-15-25-4-01-AM").Return(nil, nil)

	today = "2025-01-01"

	provider := &songProvider{
		getter: mockGetter,
	}

	actualSong, actualSource, err := provider.ListSongs(t.Context(), domain.KRUISourceType)
	require.NoError(t, err)

	require.Len(t, actualSong, 15)
	assert.Equal(t, "Nic Cowan", actualSong[0].Artist())
	assert.Equal(t, "Sun Dress", actualSong[0].Track())
	assert.Equal(t, "Hard Headed", actualSong[0].Album())

	require.Len(t, actualSource, 15)
	assert.Equal(t, "430646780", actualSource[0].SourceID())
	assert.Equal(t, actualSong[0].SongHash(), actualSource[0].SongHash())
	assert.Equal(t, domain.KRUISourceType, actualSource[0].SourceType())
	// TODO: strip out extra spaces here...
	assert.Equal(t, "89.7FM KRUI, Iowa                                    City 11/15/25, 9:01 AM", actualSource[0].ProgramName())
	assert.Equal(t, "2025-01-01", actualSource[0].Day())

	expectedEndTime, err := time.Parse(dateformat.YearMonthDayKitchen, "2025-01-01 9:31AM")
	require.NoError(t, err)
	assert.Equal(t, expectedEndTime, actualSource[0].EndTime())
}
