package store

import (
	"context"
	"testing"

	"dinosaur/internal/history"
	"dinosaur/internal/media"
	"dinosaur/internal/playback"
)

func cleanupFixture(t *testing.T) (*Store, int64, int64) {
	t.Helper()
	db, user, other := historyFixture(t)
	ctx := context.Background()
	for _, name := range []string{"Movies", "Keep"} {
		lib, err := db.SaveLibrary(ctx, Library{Name: name, Type: "movie", RootPath: t.TempDir()})
		if err != nil {
			t.Fatal(err)
		}
		if err := db.SaveMedia(ctx, media.Item{LibraryID: lib.ID, RelativePath: "video.mp4", Title: name, Duration: 100}, 1); err != nil {
			t.Fatal(err)
		}
	}
	for _, owner := range []int64{user, other} {
		for _, id := range []int64{1, 2} {
			if err := db.SavePlaybackProgress(ctx, owner, playback.Progress{MediaID: id, Position: 100, Watched: true}); err != nil {
				t.Fatal(err)
			}
		}
	}
	id := int64(1)
	for _, source := range []string{"manual", "douban"} {
		_, err := db.CreateHistory(ctx, user, history.Entry{MediaID: &id, Title: source, Source: source, SourceID: "123", WatchedOn: "2024-04-21"})
		if err != nil {
			t.Fatal(err)
		}
	}
	return db, user, other
}

func assertCleanup(t *testing.T, db *Store, user, other int64) {
	t.Helper()
	var count int
	if err := db.DB.QueryRow(`SELECT count(*) FROM playback_progress WHERE user_id=? AND media_id=1`, user).Scan(&count); err != nil || count != 0 {
		t.Fatal("disabled progress retained", err)
	}
	if err := db.DB.QueryRow(`SELECT count(*) FROM playback_progress WHERE user_id<>? OR media_id<>1`, user).Scan(&count); err != nil || count != 3 {
		t.Fatal("unrelated progress deleted", err)
	}
	page, err := db.HistoryPage(context.Background(), user, 1, 24)
	if err != nil || page.Total != 3 {
		t.Fatalf("history cleanup scope: %+v %v", page, err)
	}
	for _, entry := range page.Items {
		if entry.Source == "local" && entry.MediaID != nil && *entry.MediaID == 1 {
			t.Fatal("disabled local history retained")
		}
	}
	page, err = db.HistoryPage(context.Background(), other, 1, 24)
	if err != nil || page.Total != 2 {
		t.Fatal("other account history deleted", err)
	}
}

func TestDisablingRecordingClearsOnlyItsAccountAndLibrary(t *testing.T) {
	db, user, other := cleanupFixture(t)
	ctx := context.Background()
	for range 2 {
		if err := db.SetHistoryLibrarySetting(ctx, user, 1, false); err != nil {
			t.Fatal(err)
		}
		assertCleanup(t, db, user, other)
	}
	if err := db.SetHistoryLibrarySetting(ctx, user, 1, true); err != nil {
		t.Fatal(err)
	}
	p, err := db.PlaybackProgress(ctx, user, 1)
	if err != nil || p.Position != 0 || p.Watched {
		t.Fatal("cleared progress returned")
	}
}

func TestCleanupMigrationHandlesPreviouslyDisabledLibraries(t *testing.T) {
	db, user, other := cleanupFixture(t)
	if _, err := db.DB.Exec(`INSERT INTO history_library_settings VALUES(?,1,0,1)`, user); err != nil {
		t.Fatal(err)
	}
	if _, err := db.DB.Exec(`DELETE FROM schema_migrations WHERE version='023_clear_disabled_recording.sql'`); err != nil {
		t.Fatal(err)
	}
	for range 2 {
		if err := db.migrate(context.Background()); err != nil {
			t.Fatal(err)
		}
		assertCleanup(t, db, user, other)
	}
}

func TestRecordingCleanupFailureRollsBackSettingAndProgress(t *testing.T) {
	db, user, _ := cleanupFixture(t)
	if _, err := db.DB.Exec(`CREATE TRIGGER reject_history_cleanup BEFORE DELETE ON watch_history BEGIN SELECT RAISE(ABORT,'test cleanup failed'); END`); err != nil {
		t.Fatal(err)
	}
	if err := db.SetHistoryLibrarySetting(context.Background(), user, 1, false); err == nil {
		t.Fatal("cleanup failure accepted")
	}
	settings, err := db.HistoryLibrarySettings(context.Background(), user)
	if err != nil || !settings[0].RecordHistory {
		t.Fatal("failed cleanup persisted setting")
	}
	p, err := db.PlaybackProgress(context.Background(), user, 1)
	if err != nil || p.Position != 100 || !p.Watched {
		t.Fatal("failed cleanup partially deleted progress")
	}
	page, err := db.HistoryPage(context.Background(), user, 1, 24)
	if err != nil || page.Total != 4 {
		t.Fatal("failed cleanup deleted history")
	}
}
