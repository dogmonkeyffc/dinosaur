package scraping

import (
	"context"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"dinosaur/internal/media"
	"dinosaur/internal/metadata"
	"dinosaur/internal/store"
)

type fakeProvider struct {
	mu         sync.Mutex
	searches   int
	details    int
	candidates []metadata.Candidate
	movie      metadata.Movie
	failure    error
	block      chan struct{}
}

func (f *fakeProvider) Search(ctx context.Context, _ metadata.Settings, _ string, _ int) ([]metadata.Candidate, error) {
	f.mu.Lock()
	f.searches++
	f.mu.Unlock()
	if f.block != nil {
		select {
		case <-f.block:
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
	return f.candidates, f.failure
}
func (f *fakeProvider) Movie(context.Context, metadata.Settings, int64) (metadata.Movie, error) {
	f.mu.Lock()
	f.details++
	f.mu.Unlock()
	return f.movie, f.failure
}
func (f *fakeProvider) CachePoster(context.Context, string, string) (string, error) {
	return "", metadata.ErrProviderUnavailable
}
func (f *fakeProvider) calls() (int, int) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.searches, f.details
}

func scrapeFixture(t *testing.T) (*Manager, *store.Store, *fakeProvider) {
	t.Helper()
	ctx := context.Background()
	db, err := store.Open(ctx, filepath.Join(t.TempDir(), "scraping.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	lib, err := db.SaveLibrary(ctx, store.Library{Name: "Movies", Type: "movie", RootPath: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.SaveMedia(ctx, media.Item{LibraryID: lib.ID, RelativePath: "Example.2024.mkv", Title: "Example", Year: 2024}, 1); err != nil {
		t.Fatal(err)
	}
	if _, err := db.UpdateMetadataSettings(ctx, metadata.SettingsUpdate{Credential: "private-test-key"}, false); err != nil {
		t.Fatal(err)
	}
	provider := &fakeProvider{
		candidates: []metadata.Candidate{{ID: 10, Title: "中文电影", OriginalTitle: "Example", Year: 2024}},
		movie:      metadata.Movie{TMDBID: 10, Title: "中文电影", OriginalTitle: "Example", Year: 2024, Status: "matched", PosterPath: "/abc.jpg"},
	}
	manager := New(db, provider, t.TempDir())
	t.Cleanup(manager.Close)
	return manager, db, provider
}

func waitScrape(t *testing.T, manager *Manager, db *store.Store, id int64) metadata.Job {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		jobs, err := db.ScrapeJobs(context.Background())
		if err != nil {
			t.Fatal(err)
		}
		manager.mu.Lock()
		idle := !manager.active
		manager.mu.Unlock()
		for _, job := range jobs {
			if job.ID == id && idle && job.Status != "running" && job.Status != "queued" {
				return job
			}
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatal("scraping did not finish")
	return metadata.Job{}
}
