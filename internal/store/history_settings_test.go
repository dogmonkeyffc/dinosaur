package store

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"dinosaur/internal/history"
	"dinosaur/internal/media"
	"dinosaur/internal/playback"
)

func TestHistoryRecordingSettingsArePerAccountAndSuppressProgress(t *testing.T) {
	db, admin, viewer := historyFixture(t)
	ctx := context.Background()
	for _, name := range []string{"Movies", "Other movies"} {
		lib, err := db.SaveLibrary(ctx, Library{Name: name, Type: "movie", RootPath: t.TempDir()})
		if err != nil {
			t.Fatal(err)
		}
		if err := db.SaveMedia(ctx, media.Item{LibraryID: lib.ID, RelativePath: "video.mp4", Title: name, Duration: 100}, 1); err != nil {
			t.Fatal(err)
		}
	}
	settings, err := db.HistoryLibrarySettings(ctx, admin)
	if err != nil || len(settings) != 2 || !settings[0].RecordHistory || !settings[1].RecordHistory {
		t.Fatalf("bad defaults: %+v %v", settings, err)
	}
	if err := db.SetHistoryLibrarySetting(ctx, admin, 1, false); err != nil {
		t.Fatal(err)
	}
	p := playback.Progress{MediaID: 1, Position: 100, Watched: true, Completed: true}
	if err := db.SavePlaybackProgress(ctx, admin, p); err != nil {
		t.Fatal(err)
	}
	progress, err := db.PlaybackProgress(ctx, admin, 1)
	if err != nil || progress.Watched || progress.Position != 0 {
		t.Fatal("disabled library exposed progress")
	}
	var saved int
	if err := db.DB.QueryRow(`SELECT count(*) FROM playback_progress WHERE user_id=? AND media_id=1`, admin).Scan(&saved); err != nil || saved != 0 {
		t.Fatal("disabled library persisted progress", err)
	}
	page, _ := db.HistoryPage(ctx, admin, 1, 24)
	if page.Total != 0 {
		t.Fatal("disabled library recorded completion")
	}
	if err := db.SavePlaybackProgress(ctx, viewer, p); err != nil {
		t.Fatal(err)
	}
	other, _ := db.HistoryPage(ctx, viewer, 1, 24)
	if other.Total != 1 {
		t.Fatal("setting leaked to another account")
	}
	p.MediaID = 2
	if err := db.SavePlaybackProgress(ctx, admin, p); err != nil {
		t.Fatal(err)
	}
	page, _ = db.HistoryPage(ctx, admin, 1, 24)
	if page.Total != 1 {
		t.Fatal("setting affected another library")
	}
	if err := db.SetHistoryLibrarySetting(ctx, admin, 1, true); err != nil {
		t.Fatal(err)
	}
	page, _ = db.HistoryPage(ctx, admin, 1, 24)
	if page.Total != 1 {
		t.Fatal("enabling history backfilled suppressed events")
	}
	p.MediaID = 1
	p.Watched, p.Completed = false, false
	if err := db.SavePlaybackProgress(ctx, admin, p); err != nil {
		t.Fatal(err)
	}
	p.Watched, p.Completed = true, true
	if err := db.SavePlaybackProgress(ctx, admin, p); err != nil {
		t.Fatal(err)
	}
	if err := db.SetHistoryLibrarySetting(ctx, admin, 1, false); err != nil {
		t.Fatal(err)
	}
	imported := history.Entry{Title: "Imported", Source: "douban", SourceID: "123", WatchedOn: "2024-04-21"}
	if _, err := db.ImportHistory(ctx, admin, []history.Entry{imported}); err != nil {
		t.Fatal(err)
	}
	page, _ = db.HistoryPage(ctx, admin, 1, 24)
	if page.Total != 2 {
		t.Fatal("setting did not clear local history selectively or blocked explicit import")
	}
	var seq int
	var name, path string
	if err := db.DB.QueryRow(`PRAGMA database_list`).Scan(&seq, &name, &path); err != nil {
		t.Fatal(err)
	}
	reopened, err := Open(ctx, path)
	if err != nil {
		t.Fatal(err)
	}
	defer reopened.Close()
	settings, err = reopened.HistoryLibrarySettings(ctx, admin)
	if err != nil || settings[0].RecordHistory {
		t.Fatal("recording preference was not persisted")
	}
}

func TestHistorySettingsRespectLibraryAccess(t *testing.T) {
	db, _, viewer := historyFixture(t)
	ctx := context.Background()
	lib, err := db.SaveLibrary(ctx, Library{Name: "Private", Type: "movie", RootPath: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.DB.Exec(`UPDATE libraries SET restricted=1 WHERE id=?`, lib.ID); err != nil {
		t.Fatal(err)
	}
	viewerCtx := WithLibraryViewer(ctx, viewer)
	items, err := db.HistoryLibrarySettings(viewerCtx, viewer)
	if err != nil || len(items) != 0 {
		t.Fatal("restricted library leaked through settings")
	}
	if err := db.SetHistoryLibrarySetting(viewerCtx, viewer, lib.ID, false); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("restricted setting update: %v", err)
	}
	if err := db.SetHistoryLibrarySetting(ctx, viewer, 9999, false); !errors.Is(err, sql.ErrNoRows) {
		t.Fatal("unknown library accepted")
	}
}
