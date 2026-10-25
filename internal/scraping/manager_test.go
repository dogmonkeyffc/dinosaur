package scraping

import (
	"context"
	"errors"
	"testing"

	"dinosaur/internal/media"
	"dinosaur/internal/metadata"
)

func TestMatchedMoviesSkipProviderUntilExplicitRefresh(t *testing.T) {
	manager, db, provider := scrapeFixture(t)
	ctx := context.Background()
	job, err := manager.Start(ctx, metadata.ScrapeRequest{})
	if err != nil {
		t.Fatal(err)
	}
	result := waitScrape(t, manager, db, job.ID)
	if result.Matched != 1 || result.Status != "completed" {
		t.Fatalf("invalid result: %+v", result)
	}
	item, err := db.MediaByID(ctx, 1)
	if err != nil || item.Title != "中文电影" || item.Metadata.Message != "poster_unavailable" {
		t.Fatal("failed artwork prevented metadata match")
	}
	for range 2 {
		if err := db.SaveMedia(ctx, media.Item{LibraryID: item.LibraryID, RelativePath: item.RelativePath, Title: "Example", Year: 2024, Size: 123}, 2); err != nil {
			t.Fatal(err)
		}
		job, err = manager.Start(ctx, metadata.ScrapeRequest{})
		if err != nil {
			t.Fatal(err)
		}
		if result := waitScrape(t, manager, db, job.ID); result.Total != 0 {
			t.Fatal("matched movie reprocessed")
		}
	}
	if searches, details := provider.calls(); searches != 1 || details != 1 {
		t.Fatal("provider was called for matched movie")
	}
	job, err = manager.Start(ctx, metadata.ScrapeRequest{MediaID: 1, Refresh: true})
	if err != nil {
		t.Fatal(err)
	}
	waitScrape(t, manager, db, job.ID)
	if searches, details := provider.calls(); searches != 1 || details != 2 {
		t.Fatal("refresh searched again instead of reusing TMDB ID")
	}
	provider.failure = metadata.ErrProviderAuth
	job, err = manager.Start(ctx, metadata.ScrapeRequest{MediaID: 1, Refresh: true})
	if err != nil {
		t.Fatal(err)
	}
	result = waitScrape(t, manager, db, job.ID)
	item, err = db.MediaByID(ctx, 1)
	if err != nil || result.Status != "failed" || result.Message != "invalid_credential" || item.Title != "中文电影" || item.Metadata.TMDBID != 10 {
		t.Fatal("failed refresh destroyed matched metadata")
	}
}

func TestAmbiguousMatchingRequiresUserChoice(t *testing.T) {
	manager, db, provider := scrapeFixture(t)
	provider.candidates = append(provider.candidates, metadata.Candidate{ID: 20, Year: 2024})
	job, err := manager.Start(context.Background(), metadata.ScrapeRequest{})
	if err != nil {
		t.Fatal(err)
	}
	result := waitScrape(t, manager, db, job.ID)
	item, err := db.MediaByID(context.Background(), 1)
	if err != nil || result.Review != 1 || item.Metadata.Status != "needs_review" || item.Title != "Example" {
		t.Fatal("ambiguous movie was automatically replaced")
	}
	provider.movie.TMDBID, provider.movie.Title = 20, "人工选择的电影"
	job, err = manager.Start(context.Background(), metadata.ScrapeRequest{MediaID: 1, TMDBID: 20})
	if err != nil {
		t.Fatal(err)
	}
	waitScrape(t, manager, db, job.ID)
	item, err = db.MediaByID(context.Background(), 1)
	if err != nil || item.Title != "人工选择的电影" || item.Metadata.TMDBID != 20 {
		t.Fatal("explicit match was not saved")
	}
	if searches, _ := provider.calls(); searches != 1 {
		t.Fatal("explicit TMDB ID unexpectedly triggered search")
	}
}

func TestBoundedWorkAndShutdownInterruptsProvider(t *testing.T) {
	manager, db, provider := scrapeFixture(t)
	provider.block = make(chan struct{})
	job, err := manager.Start(context.Background(), metadata.ScrapeRequest{})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := manager.Start(context.Background(), metadata.ScrapeRequest{}); !errors.Is(err, ErrRunning) {
		t.Fatal("concurrent job was accepted")
	}
	manager.Close()
	result := waitScrape(t, manager, db, job.ID)
	if result.Status != "interrupted" {
		t.Fatal("cancellation did not preserve terminal job status")
	}
	if _, err := manager.Start(context.Background(), metadata.ScrapeRequest{}); !errors.Is(err, ErrClosed) {
		t.Fatal("closed manager accepted work")
	}
}

func TestScanEventsCoalesceAndProcessNewFiles(t *testing.T) {
	manager, db, provider := scrapeFixture(t)
	provider.block = make(chan struct{})
	manager.Schedule()
	jobs, err := db.ScrapeJobs(context.Background())
	if err != nil || len(jobs) != 1 {
		t.Fatal("automatic job was not created")
	}
	for range 3 {
		manager.Schedule()
	}
	if err := db.SaveMedia(context.Background(), media.Item{LibraryID: 1, RelativePath: "New.2024.mkv", Title: "New", Year: 2024}, 1); err != nil {
		t.Fatal(err)
	}
	close(provider.block)
	result := waitScrape(t, manager, db, jobs[0].ID+1)
	if result.Total != 1 || result.Matched != 1 {
		t.Fatalf("coalesced job did not select only the new file: %+v", result)
	}
	if searches, details := provider.calls(); searches != 2 || details != 2 {
		t.Fatal("scan events repeated identification of matched files")
	}
}
