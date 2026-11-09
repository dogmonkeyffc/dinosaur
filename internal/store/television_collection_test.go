package store

import (
	"context"
	"path/filepath"
	"testing"

	"dinosaur/internal/media"
	"dinosaur/internal/metadata"
	"dinosaur/internal/playback"
	"dinosaur/internal/television"
)

func TestSeasonCollectionMigrationKeepsEpisodesAndProgress(t *testing.T) {
	ctx := context.Background()
	db, err := Open(ctx, filepath.Join(t.TempDir(), "tv.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	lib, err := db.SaveLibrary(ctx, Library{Name: "TV", Type: "tv", RootPath: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	for i, path := range []string{
		"IT狂人S01-S05+SP.The.IT.Crowd.2006/IT狂人S01.The.IT.Crowd.2006/The.IT.Crowd.S01E01.mkv",
		"IT狂人S01-S05+SP.The.IT.Crowd.2006/IT狂人S02.The.IT.Crowd.2007/The.IT.Crowd.S02E01.mkv",
	} {
		name, ok := television.ParseEpisode(path)
		if !ok || !name.Collection {
			t.Fatalf("collection not recognized: %+v", name)
		}
		name.Key = "folder:" + filepath.Dir(path)
		name.Series = "旧季度标题"
		if err := db.SaveEpisode(ctx, media.Item{LibraryID: lib.ID, RelativePath: path, Title: name.Title, Duration: 100}, name, 1); err != nil {
			t.Fatal(err)
		}
		if err := (TelevisionRepository{db}).SaveMovieMetadata(ctx, int64(i+1), metadata.Movie{Title: "旧季度标题", Overview: "旧简介", Status: "matched", PosterKey: "poster"}); err != nil {
			t.Fatal(err)
		}
	}
	user, err := db.CreateUser(ctx, "viewer", "hash", "user", false)
	if err != nil {
		t.Fatal(err)
	}
	if err := db.SavePlaybackProgress(ctx, user.ID, playback.Progress{MediaID: 2, Position: 32}); err != nil {
		t.Fatal(err)
	}
	if _, err := db.DB.ExecContext(ctx, `DELETE FROM schema_migrations WHERE version='014_series_collections'`); err != nil {
		t.Fatal(err)
	}
	for range 2 {
		if err := db.migrate(ctx); err != nil {
			t.Fatal(err)
		}
	}
	page, err := db.SeriesPage(ctx, media.Filter{Page: 1, PageSize: 24}, user.ID)
	if err != nil || len(page.Items) != 1 {
		t.Fatalf("series = %+v, err = %v", page, err)
	}
	series := page.Items[0]
	if series.ID != 1 || series.Title != "IT狂人" || series.SeasonCount != 2 || series.PosterKey != "poster" || series.Metadata == nil || series.Metadata.Overview != "旧简介" {
		t.Fatalf("collection = %+v", series)
	}
	episodes, err := db.SeriesEpisodes(ctx, series.ID, user.ID)
	if err != nil || len(episodes) != 2 || episodes[1].ID != 2 || episodes[1].SeasonNumber != 2 || episodes[1].Position != 32 {
		t.Fatalf("episodes = %+v, err = %v", episodes, err)
	}
}
