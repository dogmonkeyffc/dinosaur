package store

import (
	"context"
	"path/filepath"
	"testing"

	"dinosaur/internal/media"
	"dinosaur/internal/metadata"
	"dinosaur/internal/playback"
)

func TestMetadataSurvivesRescanAndKeepsProgress(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "media.db")
	db, err := Open(ctx, path)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { db.Close() }()
	library, err := db.SaveLibrary(ctx, Library{Name: "Movies", Type: "movie", RootPath: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	file := media.Item{LibraryID: library.ID, RelativePath: "The.Chorus.2004.mkv", Title: "The Chorus", Year: 2004, Size: 100, Duration: 1000}
	if err := db.SaveMedia(ctx, file, 1); err != nil {
		t.Fatal(err)
	}
	page, err := db.MediaPage(ctx, media.Filter{Page: 1, PageSize: 24})
	if err != nil || len(page.Items) != 1 {
		t.Fatal("missing fixture")
	}
	id := page.Items[0].ID
	u, err := db.CreateUser(ctx, "viewer", "test-hash", "user", false)
	if err != nil {
		t.Fatal(err)
	}
	if err := db.SavePlaybackProgress(ctx, u.ID, playback.Progress{MediaID: id, Position: 120}); err != nil {
		t.Fatal(err)
	}
	info := metadata.Movie{TMDBID: 1, Title: "放牛班的春天", OriginalTitle: "Les Choristes", Aliases: "The Chorus\n唱诗班", Year: 2004, Status: "matched", Overview: "中文简介"}
	if err := db.SaveMovieMetadata(ctx, id, info); err != nil {
		t.Fatal(err)
	}
	file.Size, file.Duration = 200, 1100
	if err := db.SaveMedia(ctx, file, 2); err != nil {
		t.Fatal(err)
	}
	if err := db.RecordScrapeFailure(ctx, id, "failed", "provider_unavailable"); err != nil {
		t.Fatal(err)
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	db, err = Open(ctx, path)
	if err != nil {
		t.Fatal(err)
	}
	item, err := db.MediaByID(ctx, id)
	if err != nil || item.Title != info.Title || item.Metadata.Status != "matched" || item.Duration != 1100 || item.Filename != "The.Chorus.2004.mkv" {
		t.Fatalf("metadata or file identity changed: %+v %v", item, err)
	}
	for _, query := range []string{"放牛班", "Les Choristes", "The Chorus", "唱诗班"} {
		result, err := db.MediaPage(ctx, media.Filter{Page: 1, PageSize: 24, Query: query})
		if err != nil || result.Total != 1 {
			t.Fatalf("search failed for %s: %v", query, err)
		}
	}
	items, err := db.ContinueWatching(ctx, u.ID)
	if err != nil || len(items) != 1 || items[0].Title != info.Title || items[0].Position != 120 {
		t.Fatal("watching progress or continue title changed")
	}
	targets, err := db.ScrapeTargets(ctx, metadata.ScrapeRequest{})
	if err != nil || len(targets) != 0 {
		t.Fatal("matched movie was scheduled again")
	}
	if err := db.DeleteLibrary(ctx, library.ID); err != nil {
		t.Fatal(err)
	}
	var count int
	if err := db.DB.QueryRowContext(ctx, `SELECT count(*) FROM movie_metadata`).Scan(&count); err != nil || count != 0 {
		t.Fatal("deleted library left metadata behind")
	}
}
