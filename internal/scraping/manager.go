// Package scraping coordinates background metadata work independently of scans.
package scraping

import (
	"context"
	"errors"
	"sync"
	"time"

	"dinosaur/internal/library"
	"dinosaur/internal/media"
	"dinosaur/internal/metadata"
)

var ErrRunning = errors.New("a scraping job is already running")
var ErrClosed = errors.New("scraper is shutting down")
var ErrRequest = errors.New("invalid scraping request")

type Repository interface {
	MetadataSettings(context.Context) (metadata.Settings, error)
	MediaByID(context.Context, int64) (media.Item, error)
	LibraryByID(context.Context, int64) (library.Library, error)
	ScrapeTargets(context.Context, metadata.ScrapeRequest) ([]media.Item, error)
	StartScrapeItems(context.Context, int64, []media.Item) error
	SaveScrapeItem(context.Context, int64, int64, string, string) error
	NewScrapeJob(context.Context, metadata.ScrapeRequest) (metadata.Job, error)
	SaveScrapeJob(context.Context, metadata.Job) error
	SaveMovieMetadata(context.Context, int64, metadata.Movie) error
	RecordScrapeFailure(context.Context, int64, string, string) error
	RecordScrapeReview(context.Context, int64, []metadata.Candidate) error
}

type Provider interface {
	Search(context.Context, metadata.Settings, string, int) ([]metadata.Candidate, error)
	Movie(context.Context, metadata.Settings, int64) (metadata.Movie, error)
	CachePoster(context.Context, string, string) (string, error)
}

type Manager struct {
	repo     Repository
	provider Provider
	dataDir  string
	ctx      context.Context
	cancel   context.CancelFunc
	mu       sync.Mutex
	active   bool
	pending  bool
	closed   bool
	workers  sync.WaitGroup
}

func New(repo Repository, provider Provider, dataDir string) *Manager {
	ctx, cancel := context.WithCancel(context.Background())
	return &Manager{repo: repo, provider: provider, dataDir: dataDir, ctx: ctx, cancel: cancel}
}

func (m *Manager) Start(ctx context.Context, request metadata.ScrapeRequest) (metadata.Job, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.startLocked(ctx, request)
}

func (m *Manager) startLocked(ctx context.Context, request metadata.ScrapeRequest) (metadata.Job, error) {
	if m.closed {
		return metadata.Job{}, ErrClosed
	}
	if m.active {
		return metadata.Job{}, ErrRunning
	}
	if request.LibraryID < 0 || request.MediaID < 0 || request.TMDBID < 0 || request.TMDBID > 2147483647 ||
		(request.Refresh && request.MediaID == 0) || (request.TMDBID > 0 && request.MediaID == 0) {
		return metadata.Job{}, ErrRequest
	}
	if request.TMDBID > 0 {
		request.Refresh = true
	}
	if err := m.validateScope(ctx, request); err != nil {
		return metadata.Job{}, err
	}
	settings, err := m.repo.MetadataSettings(ctx)
	if err != nil {
		return metadata.Job{}, err
	}
	if settings.Credential == "" {
		return metadata.Job{}, metadata.ErrMissingCredential
	}
	items, err := m.repo.ScrapeTargets(ctx, request)
	if err != nil {
		return metadata.Job{}, err
	}
	job, err := m.repo.NewScrapeJob(ctx, request)
	if err != nil {
		return job, err
	}
	if err := m.repo.StartScrapeItems(ctx, job.ID, items); err != nil {
		job.Status, job.Message, job.EndedAt = "failed", "storage_error", time.Now().Unix()
		recovery, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = m.repo.SaveScrapeJob(recovery, job)
		return job, err
	}
	job.Total = len(items)
	m.active = true
	m.workers.Add(1)
	go func() {
		defer m.workers.Done()
		m.run(job, request, settings, items)
		m.mu.Lock()
		m.active = false
		pending := m.pending && !m.closed
		m.pending = false
		m.mu.Unlock()
		if pending {
			m.Schedule()
		}
	}()
	return job, nil
}

func (m *Manager) validateScope(ctx context.Context, request metadata.ScrapeRequest) error {
	libraryID := request.LibraryID
	if request.MediaID > 0 {
		item, err := m.repo.MediaByID(ctx, request.MediaID)
		if err != nil {
			return err
		}
		if libraryID > 0 && item.LibraryID != libraryID {
			return ErrRequest
		}
		libraryID = item.LibraryID
	}
	if libraryID > 0 {
		item, err := m.repo.LibraryByID(ctx, libraryID)
		if err != nil {
			return err
		}
		if item.Type != "movie" {
			return ErrRequest
		}
	}
	return nil
}

// Schedule coalesces scan-completion events into a follow-up unmatched-only job.
func (m *Manager) Schedule() {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.closed {
		return
	}
	if m.active {
		m.pending = true
		return
	}
	_, _ = m.startLocked(m.ctx, metadata.ScrapeRequest{})
}

func (m *Manager) Close() {
	m.mu.Lock()
	m.closed = true
	m.cancel()
	m.mu.Unlock()
	m.workers.Wait()
}
