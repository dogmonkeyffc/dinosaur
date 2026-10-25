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

func tvFixture(t *testing.T) (*Store, Library, []media.Item) {
	t.Helper()
	ctx := context.Background()
	db, err := Open(ctx, filepath.Join(t.TempDir(), "tv.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	lib, err := db.SaveLibrary(ctx, Library{Name: "TV", Type: "tv", RootPath: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{"Example (2020)/Season 1/Example.S01E01-E02.mp4", "Example (2020)/Season 1/Example.S01E03.mp4", "Example (2020)/Season 2/Example.S02E01.mp4"} {
		episode, ok := television.ParseEpisode(path)
		if !ok {
			t.Fatal("invalid fixture")
		}
		item := media.Item{LibraryID: lib.ID, RelativePath: path, Title: episode.Title, Duration: 100, VideoCodec: "h264"}
		if err := db.SaveEpisode(ctx, item, episode, 1); err != nil {
			t.Fatal(err)
		}
	}
	items, err := db.SeriesEpisodes(ctx, 1, 0)
	if err != nil || len(items) != 3 {
		t.Fatalf("episodes=%+v %v", items, err)
	}
	return db, lib, items
}

func TestTVIdentityProgressAndMissingIsolation(t *testing.T) {
	db, lib, items := tvFixture(t)
	ctx := context.Background()
	user, err := db.CreateUser(ctx, "viewer", "hash", "user", false)
	if err != nil {
		t.Fatal(err)
	}
	page, err := db.SeriesPage(ctx, media.Filter{Page: 1, PageSize: 24}, user.ID)
	if err != nil || page.Total != 1 || page.Items[0].SeasonCount != 2 || page.Items[0].EpisodeCount != 3 {
		t.Fatalf("series=%+v %v", page, err)
	}
	movies, err := db.MediaPage(ctx, media.Filter{Page: 1, PageSize: 24})
	if err != nil || movies.Total != 0 {
		t.Fatal("episodes leaked into movies")
	}
	next, err := db.NextEpisode(ctx, items[0])
	if err != nil || next == nil || next.EpisodeNumber != 3 {
		t.Fatalf("multi episode next=%+v %v", next, err)
	}
	next, err = db.NextEpisode(ctx, items[1])
	if err != nil || next == nil || next.SeasonNumber != 2 {
		t.Fatal("cross-season next missing")
	}
	p := playback.Progress{MediaID: items[0].ID, Position: 100, Watched: true}
	if err := db.SavePlaybackProgress(ctx, user.ID, p); err != nil {
		t.Fatal(err)
	}
	continuing, err := db.ContinueSeries(ctx, user.ID)
	if err != nil || len(continuing) != 1 || continuing[0].ID != items[1].ID {
		t.Fatalf("continue=%+v %v", continuing, err)
	}
	other, _ := db.CreateUser(ctx, "other", "hash", "user", false)
	continuing, err = db.ContinueSeries(ctx, other.ID)
	if err != nil || len(continuing) != 0 {
		t.Fatal("progress leaked across users")
	}
	repo := TelevisionRepository{db}
	if err := repo.SaveMovieMetadata(ctx, 1, metadata.Movie{TMDBID: 123, Title: "中文剧名", Aliases: "Another alias", Status: "matched"}); err != nil {
		t.Fatal(err)
	}
	for _, item := range items {
		episode, _ := television.ParseEpisode(item.RelativePath)
		if err := db.SaveEpisode(ctx, item, episode, 2); err != nil {
			t.Fatal(err)
		}
	}
	refreshed, err := db.MediaByID(ctx, items[0].ID)
	if err != nil || refreshed.SeriesTitle != "中文剧名" || refreshed.ID != items[0].ID {
		t.Fatal("rescan lost series identity or metadata")
	}
	page, err = db.SeriesPage(ctx, media.Filter{Page: 1, PageSize: 24, Query: "Another alias"}, user.ID)
	if err != nil || page.Total != 1 || page.Items[0].WatchedCount != 1 {
		t.Fatal("localized search or watch count failed")
	}
	if err := db.MarkMissingMedia(ctx, lib.ID, 999); err != nil {
		t.Fatal(err)
	}
	page, err = db.SeriesPage(ctx, media.Filter{Page: 1, PageSize: 24}, user.ID)
	if err != nil || page.Total != 0 {
		t.Fatal("empty series remained visible")
	}
	// Returning files recover their existing IDs and progress.
	if err := db.TouchMedia(ctx, items[0].ID, 3); err != nil {
		t.Fatal(err)
	}
	found, err := db.MediaByID(ctx, items[0].ID)
	if err != nil || found.SeriesID != 1 {
		t.Fatal("returning episode lost identity")
	}
	lib.Type = "movie"
	if _, err := db.SaveLibrary(ctx, lib); err != nil {
		t.Fatal(err)
	}
	var count int
	if err := db.DB.QueryRow(`SELECT count(*) FROM tv_series`).Scan(&count); err != nil || count != 0 {
		t.Fatal("type change retained old series")
	}
}

func TestTVMetadataJobsAndSnapshots(t *testing.T) {
	db, lib, items := tvFixture(t)
	ctx := context.Background()
	repo := TelevisionRepository{db}
	job, err := repo.NewScrapeJob(ctx, metadata.ScrapeRequest{LibraryID: lib.ID})
	if err != nil || job.MediaType != "tv" {
		t.Fatal(err)
	}
	targets, err := repo.ScrapeTargets(ctx, metadata.ScrapeRequest{})
	if err != nil || len(targets) != 1 {
		t.Fatal("scraping scheduled once per episode")
	}
	if err := repo.StartScrapeItems(ctx, job.ID, targets); err != nil {
		t.Fatal(err)
	}
	info := metadata.Movie{TMDBID: 10, Title: "剧名", Status: "matched"}
	if err := repo.SaveMovieMetadata(ctx, 1, info); err != nil {
		t.Fatal(err)
	}
	if err := repo.SaveScrapeItem(ctx, job.ID, 1, "matched", ""); err != nil {
		t.Fatal(err)
	}
	if err := repo.RecordScrapeFailure(ctx, 1, "failed", "provider_unavailable"); err != nil {
		t.Fatal(err)
	}
	matched, err := db.SeriesByID(ctx, 1)
	if err != nil || matched.Metadata.TMDBID != 10 || matched.Metadata.Status != "matched" {
		t.Fatal("refresh failure erased match")
	}
	detail, err := db.ScrapeJobDetail(ctx, job.ID, metadata.ReportFilter{Page: 1, PageSize: 20})
	if err != nil || len(detail.Items) != 1 || detail.Items[0].SeriesID != 1 || detail.Items[0].MediaID != 0 || detail.Items[0].Title != "剧名" {
		t.Fatalf("snapshot=%+v %v", detail, err)
	}
	for _, item := range items {
		var count int
		if err := db.DB.QueryRow(`SELECT count(*) FROM movie_metadata WHERE media_id=?`, item.ID).Scan(&count); err != nil || count != 0 {
			t.Fatal("TV metadata overwrote movie metadata")
		}
	}
	if err := db.DeleteLibrary(ctx, lib.ID); err != nil {
		t.Fatal(err)
	}
	detail, err = db.ScrapeJobDetail(ctx, job.ID, metadata.ReportFilter{Page: 1, PageSize: 20})
	if err != nil || detail.Items[0].SeriesID != 0 || detail.Items[0].Title != "剧名" {
		t.Fatalf("historical series result lost: %v", err)
	}
}
