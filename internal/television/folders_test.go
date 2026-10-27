package television

import (
	"reflect"
	"testing"

	"dinosaur/internal/media"
)

func TestEpisodeFolders(t *testing.T) {
	episodes := []media.Item{
		{RelativePath: "示例剧集/Season 02/Show.S02E01.mkv"},
		{RelativePath: "示例剧集/Season 01/Show.S01E01.mkv"},
		{RelativePath: "示例剧集/Season 01/Show.S01E02.mkv"},
		{RelativePath: "Show.S00E01.mkv"},
		{RelativePath: "旧目录/Show.S01E01.mkv", Missing: true},
		{},
	}
	want := []string{".", "示例剧集/Season 01", "示例剧集/Season 02"}
	if got := EpisodeFolders(episodes); !reflect.DeepEqual(got, want) {
		t.Fatalf("folders = %q, want %q", got, want)
	}
	if got := EpisodeFolders(nil); got == nil || len(got) != 0 {
		t.Fatalf("empty folders = %v, want a non-nil empty list", got)
	}
}
