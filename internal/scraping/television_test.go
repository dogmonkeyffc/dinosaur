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

func TestTelevisionScrapeOncePerSeriesAndManualCorrection(t *testing.T) {
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
	for _, path := range []string{"Example (2020)/Example.S01E01.mp4", "Example (2020)/Example.S01E02.mp4"} {
		name, _ := television.ParseEpisode(path)
		if err := db.SaveEpisode(ctx, media.Item{LibraryID: lib.ID, RelativePath: path, Title: name.Title}, name, 1); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := db.UpdateMetadataSettings(ctx, metadata.SettingsUpdate{Credential: "test-placeholder-key"}, false); err != nil {
		t.Fatal(err)
	}
	provider := &fakeProvider{candidates: []metadata.Candidate{{ID: 10, Title: "Example", Year: 2020}}, movie: metadata.Movie{TMDBID: 10, Title: "中文剧名", Status: "matched"}}
	repo := store.TelevisionRepository{Store: db}
	manager := NewTelevision(repo, provider, t.TempDir())
	defer manager.Close()
	job, err := manager.Start(ctx, metadata.ScrapeRequest{LibraryID: lib.ID})
	if err != nil {
		t.Fatal(err)
	}
	completed := waitScrape(t, manager, db, job.ID)
	if completed.MediaType != "tv" || completed.Total != 1 || completed.Matched != 1 {
		t.Fatalf("job=%+v", completed)
	}
	if searches, details := provider.calls(); searches != 1 || details != 1 {
		t.Fatal("series fetched per episode")
	}
	second, err := manager.Start(ctx, metadata.ScrapeRequest{})
	if err != nil {
		t.Fatal(err)
	}
	completed = waitScrape(t, manager, db, second.ID)
	if completed.Total != 0 {
		t.Fatal("ordinary scan refetched matched series")
	}
	provider.mu.Lock()
	provider.movie = metadata.Movie{TMDBID: 20, Title: "更正剧名", Status: "matched"}
	provider.mu.Unlock()
	job, err = manager.Start(ctx, metadata.ScrapeRequest{MediaID: 1, TMDBID: 20})
	if err != nil {
		t.Fatal(err)
	}
	waitScrape(t, manager, db, job.ID)
	item, err := db.SeriesByID(ctx, 1)
	if err != nil || item.Metadata.TMDBID != 20 {
		t.Fatal("manual correction lost")
	}
	detail, err := db.ScrapeJobDetail(ctx, completed.ID, metadata.ReportFilter{Page: 1, PageSize: 20})
	if err != nil || detail.Total != 0 {
		t.Fatal("empty job gained snapshots")
	}
	movieLibrary, err := db.SaveLibrary(ctx, store.Library{Name: "Movies", Type: "movie", RootPath: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := manager.Start(ctx, metadata.ScrapeRequest{LibraryID: movieLibrary.ID}); err != ErrRequest {
		t.Fatal("TV worker accepted movie library")
	}
}
