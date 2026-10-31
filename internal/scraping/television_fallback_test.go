package scraping

import (
	"context"
	"path/filepath"
	"testing"

	"dinosaur/internal/media"
	"dinosaur/internal/metadata"
	"dinosaur/internal/store"
	"dinosaur/internal/television"
)

func TestTelevisionFallbackRetainsIdentityAndOriginalLanguage(t *testing.T) {
	ctx := context.Background()
	db, err := store.Open(ctx, filepath.Join(t.TempDir(), "tv.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	lib, err := db.SaveLibrary(ctx, store.Library{Name: "TV", Type: "tv", RootPath: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{"Example (2024)/Example.S01E01.mkv", "Example (2024)/Example.S01E02.mkv"} {
		name, _ := television.ParseEpisode(path)
		if err := db.SaveEpisode(ctx, media.Item{LibraryID: lib.ID, RelativePath: path, Title: name.Title}, name, 1); err != nil {
			t.Fatal(err)
		}
	}
	repo := store.TelevisionRepository{Store: db}
	fallback := fallbackFixture()
	m := NewTelevision(repo, &fakeProvider{}, t.TempDir()).WithFallback(secondaryProvider{fallback})
	defer m.Close()
	job, err := m.Start(ctx, metadata.ScrapeRequest{})
	if err != nil {
		t.Fatal(err)
	}
	if completed := waitScrape(t, m, db, job.ID); completed.Total != 1 || completed.Matched != 1 {
		t.Fatalf("%+v", completed)
	}
	item, err := db.SeriesByID(ctx, 1)
	if err != nil || item.Metadata.BangumiID != 10 || item.Metadata.TMDBID != 0 {
		t.Fatalf("series: %+v %v", item, err)
	}
	episode, err := db.MediaByID(ctx, 1)
	if err != nil || episode.OriginalLanguage != "ja" {
		t.Fatal("series original language did not reach playback")
	}
	snapshot, err := db.ScrapeJobDetail(ctx, job.ID, metadata.ReportFilter{Page: 1, PageSize: 20})
	if err != nil || snapshot.Items[0].BangumiID != 10 || snapshot.Items[0].SeriesID != 1 {
		t.Fatalf("snapshot: %+v %v", snapshot, err)
	}
	if err := repo.RecordScrapeReview(ctx, 1, nil); err != nil {
		t.Fatal(err)
	}
	if err := repo.RecordScrapeFailure(ctx, 1, "failed", "provider_unavailable"); err != nil {
		t.Fatal(err)
	}
	item, err = db.SeriesByID(ctx, 1)
	if err != nil || item.Metadata.BangumiID != 10 || item.Metadata.Status != "matched" {
		t.Fatal("failed refresh/review erased identified series")
	}
	targets, err := repo.ScrapeTargets(ctx, metadata.ScrapeRequest{})
	if err != nil || len(targets) != 0 {
		t.Fatal("matched Bangumi series was queued again")
	}
	if _, err := db.UpdateMetadataSettings(ctx, metadata.SettingsUpdate{Credential: "test-tmdb"}, false); err != nil {
		t.Fatal(err)
	}
	m.provider = &fakeProvider{movie: metadata.Movie{TMDBID: 20, Title: "TMDB 剧名", Status: "matched"}}
	changed, err := m.Start(ctx, metadata.ScrapeRequest{MediaID: 1, TMDBID: 20})
	if err != nil {
		t.Fatal(err)
	}
	waitScrape(t, m, db, changed.ID)
	switched, err := db.ScrapeJobDetail(ctx, changed.ID, metadata.ReportFilter{Page: 1, PageSize: 20})
	if err != nil || switched.Items[0].TMDBID != 20 || switched.Items[0].BangumiID != 0 {
		t.Fatalf("snapshot retained previous source after manual correction: %+v %v", switched, err)
	}
}
