package scraping

import (
	"context"
	"errors"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"dinosaur/internal/media"
	"dinosaur/internal/metadata"
	"dinosaur/internal/store"
	"dinosaur/internal/television"
)

type priorityProvider struct {
	started chan struct{}
	release chan struct{}
	once    sync.Once
	mu      sync.Mutex
	ids     []int64
	bangumi bool
}

func (p *priorityProvider) Search(ctx context.Context, _ metadata.Settings, _ string, _ int) ([]metadata.Candidate, error) {
	p.once.Do(func() { close(p.started) })
	select {
	case <-p.release:
	case <-ctx.Done():
		return nil, ctx.Err()
	}
	return []metadata.Candidate{{ID: 10, Title: "Example", Year: 2024}}, nil
}
func (p *priorityProvider) Movie(_ context.Context, _ metadata.Settings, id int64) (metadata.Movie, error) {
	p.mu.Lock()
	p.ids = append(p.ids, id)
	p.mu.Unlock()
	info := metadata.Movie{TMDBID: id, Title: "Selected", Year: 2024, Status: "matched"}
	if p.bangumi {
		info.TMDBID, info.BangumiID = 0, id
	}
	return info, nil
}
func (p *priorityProvider) CachePoster(context.Context, string, string) (string, error) {
	return "", nil
}

func priorityFixture(t *testing.T, kind string) (*Manager, *store.Store, Repository, *priorityProvider) {
	t.Helper()
	ctx := context.Background()
	db, err := store.Open(ctx, filepath.Join(t.TempDir(), "priority.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	lib, err := db.SaveLibrary(ctx, store.Library{Name: "Media", Type: kind, RootPath: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"Example", "Next"} {
		if kind == "tv" {
			path := name + " (2024)/" + name + ".S01E01.mkv"
			episode, _ := television.ParseEpisode(path)
			err = db.SaveEpisode(ctx, media.Item{LibraryID: lib.ID, RelativePath: path, Title: name}, episode, 1)
		} else {
			err = db.SaveMedia(ctx, media.Item{LibraryID: lib.ID, RelativePath: name + ".2024.mkv", Title: name, Year: 2024}, 1)
		}
		if err != nil {
			t.Fatal(err)
		}
	}
	if _, err := db.UpdateMetadataSettings(ctx, metadata.SettingsUpdate{Credential: "test-key"}, false); err != nil {
		t.Fatal(err)
	}
	var repo Repository = db
	if kind == "tv" {
		repo = store.TelevisionRepository{Store: db}
	}
	provider := &priorityProvider{started: make(chan struct{}), release: make(chan struct{})}
	manager := New(repo, provider, t.TempDir())
	manager.libraryType = kind
	t.Cleanup(manager.Close)
	return manager, db, repo, provider
}

func TestManualChoicePrioritizesAndSurvivesInFlightBatch(t *testing.T) {
	for _, kind := range []string{"movie", "tv"} {
		for _, source := range []string{"tmdb", "bangumi"} {
			t.Run(kind+"/"+source, func(t *testing.T) {
				m, db, repo, p := priorityFixture(t, kind)
				fallback := &priorityProvider{bangumi: true}
				m.WithFallback(fallback)
				batch, err := m.Start(context.Background(), metadata.ScrapeRequest{})
				if err != nil {
					t.Fatal(err)
				}
				select {
				case <-p.started:
				case <-time.After(time.Second):
					t.Fatal("batch did not start")
				}
				var jobs []metadata.Job
				for i, target := range []int64{2, 1} {
					req := metadata.ScrapeRequest{MediaID: target, TMDBID: int64(99 + i)}
					if source == "bangumi" {
						req.TMDBID, req.BangumiID = 0, req.TMDBID
					}
					job, err := m.Start(context.Background(), req)
					if err != nil || job.Status != "queued" {
						t.Fatalf("manual choice rejected: %+v %v", job, err)
					}
					jobs = append(jobs, job)
				}
				if _, err := m.Start(context.Background(), metadata.ScrapeRequest{}); !errors.Is(err, ErrRunning) {
					t.Fatal("another bulk job accepted")
				}
				close(p.release)
				result := waitScrape(t, m, db, batch.ID)
				if result.Matched != 1 || result.Skipped != 1 {
					t.Fatalf("batch overwrote queued choice: %+v", result)
				}
				for i, target := range []int64{2, 1} {
					item, err := repo.MediaByID(context.Background(), target)
					if err != nil {
						t.Fatal(err)
					}
					gotSource, id := item.Metadata.Identity()
					if gotSource != source || id != int64(99+i) {
						t.Fatalf("manual choice lost: %+v", item.Metadata)
					}
					if result := waitScrape(t, m, db, jobs[i].ID); result.Matched != 1 {
						t.Fatalf("manual job: %+v", result)
					}
				}
				p.mu.Lock()
				defer p.mu.Unlock()
				fallback.mu.Lock()
				defer fallback.mu.Unlock()
				if source == "tmdb" {
					if len(p.ids) != 3 || p.ids[0] != 10 || p.ids[1] != 99 || p.ids[2] != 100 {
						t.Fatalf("priority order: %v", p.ids)
					}
				} else if len(p.ids) != 1 || len(fallback.ids) != 2 || fallback.ids[0] != 99 || fallback.ids[1] != 100 {
					t.Fatalf("priority order: %v %v", p.ids, fallback.ids)
				}
			})
		}
	}
}

func TestShutdownFinishesQueuedManualJobsAsInterrupted(t *testing.T) {
	m, db, _, p := priorityFixture(t, "movie")
	batch, err := m.Start(context.Background(), metadata.ScrapeRequest{})
	if err != nil {
		t.Fatal(err)
	}
	select {
	case <-p.started:
	case <-time.After(time.Second):
		t.Fatal("batch did not start")
	}
	manual, err := m.Start(context.Background(), metadata.ScrapeRequest{MediaID: 2, TMDBID: 99})
	if err != nil {
		t.Fatal(err)
	}
	m.Close()
	for _, id := range []int64{batch.ID, manual.ID} {
		if job := waitScrape(t, m, db, id); job.Status != "interrupted" {
			t.Fatalf("orphaned manual job: %+v", job)
		}
	}
}
