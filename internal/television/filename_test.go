package television

import "testing"

func TestEpisodeFilenameAndSeriesIdentity(t *testing.T) {
	cases := []struct {
		path, key, name, title     string
		year, season, episode, end int
	}{
		{"Example (2020)/Season 01/Example.S01E02.Pilot.1080p.mkv", "folder:Example (2020)", "Example", "S01E02 · Pilot", 2020, 1, 2, 2},
		{"Example (2020)/Season 02/Example.S02E01.mkv", "folder:Example (2020)", "Example", "S02E01", 2020, 2, 1, 1},
		{"Example.2020.S01E03.1080p.WEB-DL.mkv", "title:example:2020", "Example", "S01E03", 2020, 1, 3, 3},
		{"Example.1x04.Name.mp4", "title:example:0", "Example", "S01E04 · Name", 0, 1, 4, 4},
		{"剧集/特别篇/剧集.S00E01.mkv", "folder:剧集", "剧集", "S00E01", 0, 0, 1, 1},
		{"Example/Season 1/S01E01-E02.mkv", "folder:Example", "Example", "S01E01–E02", 0, 1, 1, 2},
	}
	for _, c := range cases {
		t.Run(c.path, func(t *testing.T) {
			result, ok := ParseEpisode(c.path)
			if !ok || result.Key != c.key || result.Series != c.name || result.Title != c.title || result.Year != c.year || result.Season != c.season || result.Episode != c.episode || result.End != c.end {
				t.Fatalf("parsed=%+v ok=%v", result, ok)
			}
		})
	}
	for _, path := range []string{"Movie.2020.mkv", "Example.S01E00.mkv", "S01E01.mkv", "Example.S01E02E01.mkv", "Example.S01E01E02E03.mkv", "Example.S01E01-03.mkv", "Example.S01E12345.mkv"} {
		if result, ok := ParseEpisode(path); ok {
			t.Errorf("ambiguous name accepted: %s %+v", path, result)
		}
	}
}
