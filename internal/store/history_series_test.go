package store

import (
	"context"
	"fmt"
	"testing"

	"dinosaur/internal/history"
	"dinosaur/internal/media"
	"dinosaur/internal/playback"
	"dinosaur/internal/television"
)

func historyEpisodes(t *testing.T, db *Store) int64 {
	t.Helper()
	ctx := context.Background()
	lib, err := db.SaveLibrary(ctx, Library{Name: "Series", Type: "tv", RootPath: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	for i, item := range []struct {
		key, title      string
		season, episode int
	}{
		{"one", "First series", 1, 1},
		{"one", "First series", 1, 2},
		{"one", "First series", 0, 1},
		{"two", "Second series", 2, 1},
		{"two", "Second series", 2, 1},
	} {
		media := media.Item{LibraryID: lib.ID, RelativePath: fmt.Sprintf("%d.mp4", i), Title: "Single episode", Duration: 100}
		name := television.EpisodeName{Key: item.key, Series: item.title, Season: item.season, Episode: item.episode, End: item.episode}
		if err := db.SaveEpisode(ctx, media, name, 0); err != nil {
			t.Fatal(err)
		}
	}
	return lib.ID
}

func TestWorkHistoryRequiresAllRegularEpisodesForTheSameUser(t *testing.T) {
	db, user, other := historyFixture(t)
	ctx := context.Background()
	library := historyEpisodes(t, db)
	work, err := db.SaveWork(ctx, 0, "Complete work", []int64{1, 2})
	if err != nil {
		t.Fatal(err)
	}
	save := func(owner, id int64, position float64, completed bool) {
		t.Helper()
		p := playback.Progress{MediaID: id, Position: position, Watched: completed, Completed: completed}
		if err := db.SavePlaybackProgress(ctx, owner, p); err != nil {
			t.Fatal(err)
		}
	}
	save(user, 1, 100, true)
	save(user, 2, 95, false)
	save(user, 4, 100, true)
	save(other, 2, 100, true)
	page, err := db.HistoryPage(ctx, user, 1, 24)
	if err != nil || page.Total != 0 {
		t.Fatalf("incomplete work logged: %+v %v", page, err)
	}
	if p, err := db.PlaybackProgress(ctx, user, 2); err != nil || p.ResumePosition() != 95 {
		t.Fatalf("incomplete episode cannot resume: %+v %v", p, err)
	}
	save(user, 2, 100, true)
	save(user, 4, 100, true)
	page, err = db.HistoryPage(ctx, user, 1, 24)
	if err != nil || page.Total != 1 || page.Items[0].WorkID == nil || *page.Items[0].WorkID != work.ID ||
		page.Items[0].MediaID != nil || page.Items[0].Title != "Complete work" {
		t.Fatalf("wrong work entry: %+v %v", page, err)
	}
	if err := db.SetHistoryLibrarySetting(ctx, user, library, false); err != nil {
		t.Fatal(err)
	}
	page, err = db.HistoryPage(ctx, user, 1, 24)
	if err != nil || page.Total != 0 {
		t.Fatalf("disabled work history retained: %+v %v", page, err)
	}
}

func TestEpisodeHistoryMigrationOnlyConsolidatesCompletedSeries(t *testing.T) {
	db, user, other := historyFixture(t)
	ctx := context.Background()
	historyEpisodes(t, db)
	for _, statement := range []string{
		`DROP INDEX watch_history_local_series`,
		`DROP INDEX watch_history_local_work`,
		`DROP VIEW history_episode_catalog`,
		`ALTER TABLE watch_history DROP COLUMN series_id`,
		`ALTER TABLE watch_history DROP COLUMN work_id`,
		`ALTER TABLE playback_progress DROP COLUMN completed_at`,
		`DELETE FROM schema_migrations WHERE version='026_work_watch_history.sql'`,
	} {
		if _, err := db.DB.Exec(statement); err != nil {
			t.Fatal(err)
		}
	}
	for _, legacy := range []struct {
		user, mediaID int64
		position      float64
		date          string
	}{
		{user, 1, 100, "2024-04-20"},
		{user, 2, 100, "2024-04-21"},
		{other, 1, 100, "2024-04-20"},
		{other, 2, 90, "2024-04-21"},
	} {
		_, err := db.DB.Exec(`INSERT INTO watch_history(user_id,media_id,source,source_id,title,watched_on,created_at,updated_at)
 VALUES(?,?,'local',?,'Single episode',?,1713657600,1713657600)`, legacy.user, legacy.mediaID, fmt.Sprint(legacy.mediaID), legacy.date)
		if err != nil {
			t.Fatal(err)
		}
		_, err = db.DB.Exec(`INSERT INTO playback_progress(user_id,media_id,position,watched,updated_at) VALUES(?,?,?,1,1)`, legacy.user, legacy.mediaID, legacy.position)
		if err != nil {
			t.Fatal(err)
		}
	}
	if _, err := db.DB.Exec(`INSERT INTO watch_history(user_id,media_id,source,source_id,title,watched_on,created_at,updated_at)
 VALUES(?,1,'douban','123','Imported','2024-04-22',1,1)`, user); err != nil {
		t.Fatal(err)
	}
	for range 2 {
		if err := db.migrate(ctx); err != nil {
			t.Fatal(err)
		}
		page, err := db.HistoryPage(ctx, user, 1, 24)
		if err != nil || page.Total != 2 || page.Items[0].Source != "douban" || page.Items[1].SeriesID == nil ||
			page.Items[1].Title != "First series" || page.Items[1].WatchedOn != "2024-04-21" {
			t.Fatalf("wrong migrated series: %+v %v", page, err)
		}
		page, err = db.HistoryPage(ctx, other, 1, 24)
		if err != nil || page.Total != 0 {
			t.Fatalf("partial series migrated: %+v %v", page, err)
		}
	}
	if _, err := db.ImportHistory(ctx, user, []history.Entry{{Title: "Imported", Source: "douban", SourceID: "123", WatchedOn: "2024-04-22"}}); err != nil {
		t.Fatal(err)
	}
	if p, err := db.PlaybackProgress(ctx, other, 2); err != nil || p.Position != 90 {
		t.Fatalf("migration changed playback progress: %+v %v", p, err)
	}
}
