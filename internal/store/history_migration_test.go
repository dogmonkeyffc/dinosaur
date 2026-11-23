package store

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"dinosaur/internal/media"
)

func TestHistoryMigrationPreservesExistingWatchedProgress(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "migration.db")
	db, err := Open(ctx, path)
	if err != nil {
		t.Fatal(err)
	}
	u, err := db.CreateUser(ctx, "admin", "hash", "admin", true)
	if err != nil {
		t.Fatal(err)
	}
	lib, err := db.SaveLibrary(ctx, Library{Name: "Movies", Type: "movie", RootPath: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.SaveMedia(ctx, media.Item{LibraryID: lib.ID, RelativePath: "old.mp4", Title: "Existing watched movie", Year: 2020}, 1); err != nil {
		t.Fatal(err)
	}
	stamp := time.Date(2024, 4, 20, 17, 0, 0, 0, time.UTC).Unix()
	for _, query := range []string{`DROP TABLE watch_history`, `DELETE FROM schema_migrations WHERE version='020_watch_history.sql'`} {
		if _, err := db.DB.Exec(query); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := db.DB.Exec(`INSERT INTO playback_progress(user_id,media_id,position,watched,updated_at) VALUES(?,1,100,1,?)`, u.ID, stamp); err != nil {
		t.Fatal(err)
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	db, err = Open(ctx, path)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	page, err := db.HistoryPage(ctx, u.ID, 1, 24)
	if err != nil || page.Total != 1 || page.Items[0].WatchedOn != "2024-04-21" || page.Items[0].Title != "Existing watched movie" {
		t.Fatalf("backfill failed: %+v %v", page, err)
	}
	if err := db.migrate(ctx); err != nil {
		t.Fatal(err)
	}
	page, _ = db.HistoryPage(ctx, u.ID, 1, 24)
	if page.Total != 1 {
		t.Fatal("migration duplicated watch history")
	}
}
