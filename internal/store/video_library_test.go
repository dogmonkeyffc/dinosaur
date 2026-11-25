package store

import (
	"context"
	"database/sql"
	"testing"

	"dinosaur/internal/media"
	"dinosaur/internal/metadata"
	"dinosaur/internal/playback"
)

func TestVideoLibraryMigrationPreservesReferencesAndEnforcement(t *testing.T) {
	ctx := context.Background()
	raw, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	raw.SetMaxOpenConns(1)
	db := &Store{DB: raw}
	defer db.Close()
	if _, err := raw.Exec(`PRAGMA foreign_keys=ON`); err != nil {
		t.Fatal(err)
	}
	if err := db.migrate(ctx); err != nil {
		t.Fatal(err)
	}
	u, err := db.CreateUser(ctx, "viewer", "hash", "user", false)
	if err != nil {
		t.Fatal(err)
	}
	lib, err := db.SaveLibrary(ctx, Library{Name: "Keep", Type: "movie", RootPath: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.SaveMedia(ctx, media.Item{LibraryID: lib.ID, RelativePath: "keep.mp4", Title: "Keep", Duration: 100}, 1); err != nil {
		t.Fatal(err)
	}
	if err := db.SavePlaybackProgress(ctx, u.ID, playback.Progress{MediaID: 1, Position: 100, Watched: true}); err != nil {
		t.Fatal(err)
	}
	if err := db.SetHistoryLibrarySetting(ctx, u.ID, lib.ID, true); err != nil {
		t.Fatal(err)
	}
	for range 2 {
		if err := db.migrateVideoLibraries(ctx); err != nil {
			t.Fatal(err)
		}
	}
	p, err := db.PlaybackProgress(ctx, u.ID, 1)
	if err != nil || !p.Watched || p.Position != 100 {
		t.Fatalf("migration lost progress: %+v %v", p, err)
	}
	page, err := db.HistoryPage(ctx, u.ID, 1, 24)
	if err != nil || page.Total != 1 {
		t.Fatal("migration lost history", err)
	}
	var enabled int
	if err := db.DB.QueryRow(`PRAGMA foreign_keys`).Scan(&enabled); err != nil || enabled != 1 {
		t.Fatal("foreign keys not restored")
	}
	if _, err := db.DB.Exec(`INSERT INTO playback_progress(user_id,media_id) VALUES(999,999)`); err == nil {
		t.Fatal("foreign key enforcement disabled")
	}
	video, err := db.SaveLibrary(ctx, Library{Name: "Videos", Type: "video", RootPath: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.SaveMedia(ctx, media.Item{LibraryID: video.ID, RelativePath: "raw.mp4", Title: "raw.mp4", Duration: 100}, 1); err != nil {
		t.Fatal(err)
	}
	movies, err := db.MediaPage(ctx, media.Filter{Page: 1, PageSize: 24})
	if err != nil || movies.Total != 1 || movies.Items[0].Type != "movie" {
		t.Fatal("video leaked into movies", err)
	}
	videos, err := db.MediaPage(ctx, media.Filter{Type: "video", Page: 1, PageSize: 24})
	if err != nil || videos.Total != 1 || videos.Items[0].Title != "raw.mp4" {
		t.Fatalf("video catalog: %+v %v", videos, err)
	}
	targets, err := db.ScrapeTargets(ctx, metadata.ScrapeRequest{Refresh: true})
	if err != nil || len(targets) != 1 || targets[0].LibraryID != lib.ID {
		t.Fatal("videos became scrape targets", err)
	}
	if err := db.SavePlaybackProgress(ctx, u.ID, playback.Progress{MediaID: 2, Position: 20}); err != nil {
		t.Fatal(err)
	}
	recent, err := db.ContinueWatching(ctx, u.ID, "video")
	if err != nil || len(recent) != 1 || recent[0].Type != "video" {
		t.Fatal("video continue missing", err)
	}
	if err := db.SetHistoryLibrarySetting(ctx, u.ID, video.ID, false); err != nil {
		t.Fatal(err)
	}
	recent, err = db.ContinueWatching(ctx, u.ID, "video")
	if err != nil || len(recent) != 0 {
		t.Fatal("video privacy preference ignored", err)
	}
	if err := db.DeleteLibrary(ctx, video.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := db.MediaByID(ctx, 2); err != sql.ErrNoRows {
		t.Fatal("library cascade broken", err)
	}
}
