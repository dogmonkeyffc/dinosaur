package scraping

import (
	"context"
	"errors"
	"testing"

	"dinosaur/internal/metadata"
	"dinosaur/internal/store"
)

type canceledStartRepository struct {
	*store.Store
	cancel context.CancelFunc
}

func (r canceledStartRepository) NewScrapeJob(ctx context.Context, request metadata.ScrapeRequest) (metadata.Job, error) {
	job, err := r.Store.NewScrapeJob(ctx, request)
	r.cancel()
	return job, err
}

func TestCanceledStartDoesNotLeaveQueuedJob(t *testing.T) {
	_, db, provider := scrapeFixture(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	manager := New(canceledStartRepository{Store: db, cancel: cancel}, provider, t.TempDir())
	defer manager.Close()
	if _, err := manager.Start(ctx, metadata.ScrapeRequest{}); !errors.Is(err, context.Canceled) {
		t.Fatalf("unexpected start error: %v", err)
	}
	jobs, err := db.ScrapeJobs(context.Background())
	if err != nil || len(jobs) != 1 || jobs[0].Status != "failed" || jobs[0].EndedAt == 0 {
		t.Fatal("canceled request left an orphaned queued job")
	}
}
