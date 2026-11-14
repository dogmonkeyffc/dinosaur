package store

import (
	"context"
	"testing"

	"dinosaur/internal/media"
	"dinosaur/internal/metadata"
	"dinosaur/internal/television"
)

func TestWorkRescanMissingMembersAndDissolvePreserveSeries(t *testing.T) {
	db, lib, episodes := tvFixture(t)
	ctx := context.Background()
	second := addWorkSeries(t, db, lib.ID, "Separate/Beta/S03E01.mp4")
	a := episodes[0].SeriesID
	if err := (TelevisionRepository{db}).SaveMovieMetadata(ctx, a, metadata.Movie{Title: "Localized title", TMDBID: 123, Overview: "Keep me", Status: "matched"}); err != nil {
		t.Fatal(err)
	}
	user := catalogUser(t, db, "work-viewer")
	catalogProgress(t, db, user, episodes[0].ID, 100, 100, true)
	work, err := db.SaveWork(ctx, 0, "Manual collection", []int64{second.SeriesID, a})
	if err != nil {
		t.Fatal(err)
	}
	if work.EpisodeCount != 4 || work.SeasonCount != 3 || work.MemberCount != 2 {
		t.Fatalf("aggregate counts=%+v", work)
	}
	for _, item := range append(episodes, second) {
		name, _ := television.ParseEpisode(item.RelativePath)
		if err := db.SaveEpisode(ctx, item, name, 2); err != nil {
			t.Fatal(err)
		}
		refreshed, err := db.MediaByID(ctx, item.ID)
		if err != nil || refreshed.ID != item.ID || refreshed.SeriesID != item.SeriesID || refreshed.RelativePath != item.RelativePath {
			t.Fatalf("scan identity changed: before=%+v after=%+v error=%v", item, refreshed, err)
		}
	}
	requireWorkMembers(t, db, work.ID, "Manual collection", second.SeriesID, a)
	series, err := db.SeriesByID(ctx, a)
	if err != nil || series.Metadata == nil || series.Metadata.TMDBID != 123 || series.Metadata.Overview != "Keep me" {
		t.Fatalf("metadata changed: %+v %v", series, err)
	}
	detail, err := db.WorkByID(ctx, work.ID, user)
	if err != nil || detail.Work.WatchedCount != 1 || len(detail.Members) != 2 {
		t.Fatalf("detail=%+v %v", detail, err)
	}
	if _, err := db.DB.Exec(`UPDATE media_items SET missing=1 WHERE id IN (SELECT media_id FROM tv_episodes WHERE series_id=?)`, second.SeriesID); err != nil {
		t.Fatal(err)
	}
	detail, err = db.WorkByID(ctx, work.ID, user)
	if err != nil || len(detail.Members) != 1 || len(detail.AllMembers) != 2 || detail.AllMembers[0].ID != second.SeriesID || detail.Work.EpisodeCount != 3 {
		t.Fatalf("missing member lost from editable membership: %+v %v", detail, err)
	}
	missingMatch, err := db.SeriesPage(ctx, media.Filter{Page: 1, PageSize: 24, Query: "Beta"}, user)
	if err != nil || missingMatch.Total != 1 || missingMatch.Items[0].WorkID != work.ID || missingMatch.Items[0].EpisodeCount != 3 {
		t.Fatalf("missing member search lost visible work: %+v %v", missingMatch, err)
	}
	if err := db.TouchMedia(ctx, second.ID, 3); err != nil {
		t.Fatal(err)
	}
	detail, err = db.WorkByID(ctx, work.ID, user)
	if err != nil || len(detail.Members) != 2 || detail.Members[0].ID != second.SeriesID {
		t.Fatalf("returning member lost membership/order: %+v %v", detail, err)
	}
	if err := db.DeleteWork(ctx, work.ID); err != nil {
		t.Fatal(err)
	}
	page, err := db.SeriesPage(ctx, media.Filter{Page: 1, PageSize: 24, LibraryID: lib.ID}, user)
	if err != nil || page.Total != 2 {
		t.Fatalf("dissolve did not restore series: %+v %v", page, err)
	}
	for _, item := range page.Items {
		if item.WorkID != 0 {
			t.Fatalf("dissolved series retains work ID: %+v", item)
		}
	}
	items, err := db.SeriesEpisodes(ctx, a, user)
	if err != nil || len(items) != 3 || items[0].ID != episodes[0].ID || !items[0].Watched {
		t.Fatalf("dissolve lost episodes/progress: %+v %v", items, err)
	}
	series, err = db.SeriesByID(ctx, a)
	if err != nil || series.Metadata == nil || series.Metadata.TMDBID != 123 {
		t.Fatalf("dissolve lost metadata: %+v %v", series, err)
	}
}

func TestWorkCatalogSearchCollapsePaginationAndRecentOrder(t *testing.T) {
	db, lib, episodes := tvFixture(t)
	ctx := context.Background()
	beta := addWorkSeries(t, db, lib.ID, "Another/Beta/S01E01.mp4")
	standalone := addWorkSeries(t, db, lib.ID, "Standalone/S01E01.mp4")
	other, err := db.SaveLibrary(ctx, Library{Name: "Elsewhere", Type: "tv", RootPath: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	addWorkSeries(t, db, other.ID, "Elsewhere/S01E01.mp4")
	work, err := db.SaveWork(ctx, 0, "Zulu anthology", []int64{beta.SeriesID, episodes[0].SeriesID})
	if err != nil {
		t.Fatal(err)
	}
	alice, bob := catalogUser(t, db, "work-alice"), catalogUser(t, db, "work-bob")
	catalogProgress(t, db, alice, beta.ID, 900, 100, true)
	catalogProgress(t, db, alice, standalone.ID, 800, 20, false)
	catalogProgress(t, db, bob, standalone.ID, 1000, 20, false)
	filter := media.Filter{Page: 1, PageSize: 1, LibraryID: lib.ID}
	page, err := db.SeriesPage(ctx, filter, alice)
	if err != nil || page.Total != 2 || len(page.Items) != 1 || page.Items[0].WorkID != work.ID || page.Items[0].Title != work.Title || page.Items[0].MemberCount != 2 || page.Items[0].EpisodeCount != 4 || page.Items[0].WatchedCount != 1 {
		t.Fatalf("aggregate first page=%+v %v", page, err)
	}
	filter.Page = 2
	page, err = db.SeriesPage(ctx, filter, alice)
	if err != nil || page.Total != 2 || len(page.Items) != 1 || page.Items[0].ID != standalone.SeriesID || page.Items[0].WorkID != 0 {
		t.Fatalf("second page=%+v %v", page, err)
	}
	filter.Page = 1
	page, err = db.SeriesPage(ctx, filter, bob)
	if err != nil || len(page.Items) != 1 || page.Items[0].ID != standalone.SeriesID || page.Items[0].WorkID != 0 {
		t.Fatalf("recent order leaked between viewers: %+v %v", page, err)
	}
	for _, query := range []string{"Zulu anthology", "Beta", "Example"} {
		filter.Query = query
		page, err = db.SeriesPage(ctx, filter, alice)
		if err != nil || page.Total != 1 || len(page.Items) != 1 || page.Items[0].WorkID != work.ID || page.Items[0].EpisodeCount != 4 {
			t.Fatalf("search %q failed to return complete aggregate: %+v %v", query, page, err)
		}
	}
	filter.Query, filter.PageSize = "", 24
	raw, err := db.RawSeriesPage(ctx, filter, alice)
	if err != nil || raw.Total != 3 || len(raw.Items) != 3 {
		t.Fatalf("raw page=%+v %v", raw, err)
	}
	for _, item := range raw.Items {
		if item.ID != standalone.SeriesID && item.WorkID != work.ID {
			t.Fatalf("raw member lacks membership: %+v", item)
		}
	}
	works, err := db.Works(ctx, other.ID, alice)
	if err != nil || len(works) != 0 {
		t.Fatalf("work library filter=%+v %v", works, err)
	}
	works, err = db.Works(ctx, lib.ID, alice)
	if err != nil || len(works) != 1 || works[0].ID != work.ID || works[0].WatchedCount != 1 {
		t.Fatalf("work listing=%+v %v", works, err)
	}
}
