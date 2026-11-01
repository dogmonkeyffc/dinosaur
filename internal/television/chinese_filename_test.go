package television

import "testing"

func TestChineseEpisodeMarkersKeepFolderIdentity(t *testing.T) {
	for _, test := range []struct {
		path, key, title string
		season, episode  int
	}{
		{"百妖谱·司府篇/[百妖谱 司府篇][ 第01集 故人归乡][2024][2160P]  - WHfoimptor.mp4", "folder:百妖谱·司府篇", "S01E01 · 故人归乡", 1, 1},
		{"百妖谱·司府篇/[百妖谱 司府篇][ 第12集 镇水有灵][2024][2160P]  - WHfoimptor.mp4", "folder:百妖谱·司府篇", "S01E12 · 镇水有灵", 1, 12},
		{"剧集/Season 4/第02集.mp4", "folder:剧集", "S04E02", 4, 2},
		{"剧集/第4季/剧集第03集.mp4", "folder:剧集", "S04E03", 4, 3},
		{"剧集/剧集第2季第01集.mp4", "folder:剧集", "S02E01", 2, 1},
	} {
		t.Run(test.path, func(t *testing.T) {
			item, ok := ParseEpisode(test.path)
			if !ok || item.Key != test.key || item.Title != test.title || item.Season != test.season || item.Episode != test.episode {
				t.Fatalf("unexpected episode: %+v %v", item, ok)
			}
		})
	}
	for _, path := range []string{"剧集/第00集.mp4", "剧集/第01集第02集.mp4", "剧集/第01-02集.mp4", "剧集/第12345集.mp4", "剧集/2024.2160P.mp4"} {
		if item, ok := ParseEpisode(path); ok {
			t.Fatalf("ambiguous episode accepted: %s %+v", path, item)
		}
	}
}
