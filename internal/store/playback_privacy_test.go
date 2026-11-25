package store

import (
	"context"
	"testing"

	"dinosaur/internal/media"
	"dinosaur/internal/playback"
)

func TestDisabledLibraryHidesSavedMovieProgressAndIgnoresWrites(t *testing.T) {
	db, items := reportFixture(t)
	ctx := context.Background()
	user := catalogUser(t, db, "viewer")
	id := items[0].ID
	catalogProgress(t, db, user, id, 123, 40, false)
	if err := db.SetHistoryLibrarySetting(ctx, user, items[0].LibraryID, false); err != nil {
		t.Fatal(err)
	}
	progress, err := db.PlaybackProgress(ctx, user, id)
	if err != nil || progress.Position != 0 || progress.Watched || progress.UpdatedAt != 0 {
		t.Fatalf("hidden progress: %+v %v", progress, err)
	}
	if err := db.AddPlaybackProgress(ctx, user, items); err != nil {
		t.Fatal(err)
	}
	for _, item := range items {
		if item.Position != 0 || item.Watched || item.LastPlayedAt != 0 {
			t.Fatal("catalog exposed progress")
		}
	}
	recent, err := db.ContinueWatching(ctx, user)
	if err != nil || len(recent) != 0 {
		t.Fatalf("continue exposed progress: %+v %v", recent, err)
	}
	if err := db.SavePlaybackProgress(ctx, user, playback.Progress{MediaID: id, Position: 100, Watched: true}); err != nil {
		t.Fatal(err)
	}
	var count int
	if err := db.DB.QueryRow(`SELECT count(*) FROM playback_progress WHERE user_id=? AND media_id=?`, user, id).Scan(&count); err != nil || count != 0 {
		t.Fatal("disabled library retained progress", err)
	}
	if err := db.SetHistoryLibrarySetting(ctx, user, items[0].LibraryID, true); err != nil {
		t.Fatal(err)
	}
	progress, err = db.PlaybackProgress(ctx, user, id)
	if err != nil || progress.Position != 0 {
		t.Fatal("cleared progress returned after enabling")
	}
}

func TestDisabledLibraryHidesSeriesResumeAndWatchedCounts(t *testing.T) {
	db, lib, items := tvFixture(t)
	ctx := context.Background()
	user := catalogUser(t, db, "viewer")
	catalogProgress(t, db, user, items[0].ID, 123, 100, true)
	catalogProgress(t, db, user, items[1].ID, 124, 40, false)
	if err := db.SetHistoryLibrarySetting(ctx, user, lib.ID, false); err != nil {
		t.Fatal(err)
	}
	episodes, err := db.SeriesEpisodes(ctx, items[0].SeriesID, user)
	if err != nil {
		t.Fatal(err)
	}
	for _, episode := range episodes {
		if episode.Position != 0 || episode.Watched || episode.LastPlayedAt != 0 {
			t.Fatal("episode progress exposed")
		}
	}
	recent, err := db.ContinueSeries(ctx, user)
	if err != nil || len(recent) != 0 {
		t.Fatalf("series continue exposed: %+v %v", recent, err)
	}
	filter := media.Filter{Page: 1, PageSize: 24}

	page, err := db.SeriesPage(ctx, filter, user)
	if err != nil || len(page.Items) != 1 || page.Items[0].WatchedCount != 0 {
		t.Fatalf("grouped series count: %+v %v", page, err)
	}
	raw, err := db.RawSeriesPage(ctx, filter, user)
	if err != nil || len(raw.Items) != 1 || raw.Items[0].WatchedCount != 0 {
		t.Fatalf("raw series count: %+v %v", raw, err)
	}
}
