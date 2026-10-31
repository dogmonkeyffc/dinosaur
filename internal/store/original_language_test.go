package store

import (
	"context"
	"testing"

	"dinosaur/internal/metadata"
)

func TestEpisodesUseSeriesOriginalLanguage(t *testing.T) {
	db, _, episodes := tvFixture(t)
	info := metadata.Movie{TMDBID: 42, Title: "中文译名", Language: "zh-CN", OriginalLanguage: "ko", Status: "matched"}
	if err := (TelevisionRepository{Store: db}).SaveMovieMetadata(context.Background(), 1, info); err != nil {
		t.Fatal(err)
	}
	for _, episode := range episodes {
		item, err := db.MediaByID(context.Background(), episode.ID)
		if err != nil || item.OriginalLanguage != "ko" {
			t.Fatalf("episode original language=%q %v", item.OriginalLanguage, err)
		}
	}
}
