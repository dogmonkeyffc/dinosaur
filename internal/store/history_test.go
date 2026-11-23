package store

import (
	"context"
	"database/sql"
	"errors"
	"path/filepath"
	"testing"

	"dinosaur/internal/history"
	"dinosaur/internal/media"
	"dinosaur/internal/playback"
)

func historyFixture(t *testing.T) (*Store, int64, int64) {
	t.Helper()
	db, err := Open(context.Background(), filepath.Join(t.TempDir(), "history.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	a, err := db.CreateUser(context.Background(), "admin", "hash", "admin", true)
	if err != nil {
		t.Fatal(err)
	}
	b, err := db.CreateUser(context.Background(), "viewer", "hash", "user", false)
	if err != nil {
		t.Fatal(err)
	}
	return db, a.ID, b.ID
}

func TestHistoryImportIsolationAndPersistence(t *testing.T) {
	db, a, b := historyFixture(t)
	ctx := context.Background()
	e := history.Entry{Title: "Imported", Source: "douban", SourceID: "123", WatchedOn: "2024-04-21", Rating: 5}
	r, err := db.ImportHistory(ctx, a, []history.Entry{e, e})
	if err != nil || r.Imported != 1 || r.Skipped != 1 {
		t.Fatalf("%+v %v", r, err)
	}
	page, err := db.HistoryPage(ctx, a, 1, 24)
	if err != nil || page.Total != 1 {
		t.Fatalf("%+v %v", page, err)
	}
	entry := page.Items[0]
	if _, err := db.HistoryByID(ctx, b, entry.ID); !errors.Is(err, sql.ErrNoRows) {
		t.Fatal("history leaked")
	}
	entry.Title = "Corrected"
	if err := db.UpdateHistory(ctx, a, entry); err != nil {
		t.Fatal(err)
	}
	if err := db.DeleteHistory(ctx, b, entry.ID); !errors.Is(err, sql.ErrNoRows) {
		t.Fatal("cross-user delete")
	}
	r, err = db.ImportHistory(ctx, a, []history.Entry{e})
	if err != nil || r.Imported != 0 || r.Skipped != 1 {
		t.Fatalf("%+v %v", r, err)
	}
	entry, err = db.HistoryByID(ctx, a, entry.ID)
	if err != nil || entry.Title != "Corrected" {
		t.Fatal("reimport overwrote edits")
	}
	e.WatchedOn = "2025-01-01"
	if _, err := db.CreateHistory(ctx, a, e); err != nil {
		t.Fatal(err)
	}
	page, err = db.HistoryPage(ctx, a, 1, 1)
	if err != nil || page.Total != 2 || page.Items[0].WatchedOn != "2025-01-01" {
		t.Fatalf("wrong order: %+v %v", page, err)
	}
	bad := e
	bad.Rating = 6
	if _, err := db.ImportHistory(ctx, b, []history.Entry{e, bad}); err == nil {
		t.Fatal("invalid import accepted")
	}
	other, _ := db.HistoryPage(ctx, b, 1, 24)
	if other.Total != 0 {
		t.Fatal("partial import committed")
	}
	if err := db.DeleteHistory(ctx, a, entry.ID); err != nil {
		t.Fatal(err)
	}
}

func TestPlaybackCompletionCreatesHistoryOnce(t *testing.T) {
	db, user, _ := historyFixture(t)
	ctx := context.Background()
	lib, err := db.SaveLibrary(ctx, Library{Name: "Movies", Type: "movie", RootPath: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.SaveMedia(ctx, media.Item{LibraryID: lib.ID, RelativePath: "movie.mp4", Title: "Movie", Duration: 100}, 1); err != nil {
		t.Fatal(err)
	}
	p := playback.Progress{MediaID: 1, Position: 50}
	if err := db.SavePlaybackProgress(ctx, user, p); err != nil {
		t.Fatal(err)
	}
	page, _ := db.HistoryPage(ctx, user, 1, 24)
	if page.Total != 0 {
		t.Fatal("unfinished viewing logged")
	}
	p.Watched = true
	for range 3 {
		if err := db.SavePlaybackProgress(ctx, user, p); err != nil {
			t.Fatal(err)
		}
	}
	page, _ = db.HistoryPage(ctx, user, 1, 24)
	if page.Total != 1 || page.Items[0].Title != "Movie" {
		t.Fatalf("wrong completion: %+v", page)
	}
	p.Watched = false
	if err := db.SavePlaybackProgress(ctx, user, p); err != nil {
		t.Fatal(err)
	}
	p.Watched = true
	if err := db.SavePlaybackProgress(ctx, user, p); err != nil {
		t.Fatal(err)
	}
	page, _ = db.HistoryPage(ctx, user, 1, 24)
	if page.Total != 1 {
		t.Fatal("same-day duplicate completion")
	}
	if _, err := db.DB.Exec(`DELETE FROM media_items WHERE id=1`); err != nil {
		t.Fatal(err)
	}
	page, _ = db.HistoryPage(ctx, user, 1, 24)
	if page.Total != 1 || page.Items[0].MediaID != nil {
		t.Fatal("deleting media removed history")
	}
}
