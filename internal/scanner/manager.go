// Package scanner coordinates bounded background jobs for media libraries.
package scanner

import (
	"context"
	"errors"
	"log/slog"
	"sync"

	"dinosaur/internal/config"
	"dinosaur/internal/library"
	"dinosaur/internal/store"
)

var ErrRunning = errors.New("library scan is already running")
var ErrUnsupported = errors.New("unsupported library type")
var ErrClosed = errors.New("scanner is shutting down")

type Manager struct {
	// AfterScan is configured before Start and never performs network work inline.
	AfterScan func()
	db        *store.Store
	config    config.Config
	logger    *slog.Logger
	ctx       context.Context
	cancel    context.CancelFunc
	mu        sync.Mutex
	active    map[int64]bool
	closed    bool
	slots     chan struct{}
	workers   sync.WaitGroup
}

func New(db *store.Store, c config.Config, logger *slog.Logger) *Manager {
	ctx, cancel := context.WithCancel(context.Background())
	return &Manager{db: db, config: c, logger: logger, ctx: ctx, cancel: cancel, active: make(map[int64]bool), slots: make(chan struct{}, 2)}
}
func (m *Manager) Start() error {
	if err := m.db.InterruptScans(m.ctx); err != nil {
		return err
	}
	libraries, err := m.db.Libraries(m.ctx)
	if err != nil {
		return err
	}
	for _, item := range libraries {
		if item.Type != "movie" && item.Type != "tv" {
			continue
		}
		if _, err := m.Scan(m.ctx, item.ID); err != nil {
			m.logger.Error("start library scan", "library_id", item.ID, "error", err)
		}
	}
	return nil
}
func (m *Manager) Scan(ctx context.Context, id int64) (library.ScanJob, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.closed {
		return library.ScanJob{}, ErrClosed
	}
	if m.active[id] {
		return library.ScanJob{}, ErrRunning
	}
	item, err := m.db.LibraryByID(ctx, id)
	if err != nil {
		return library.ScanJob{}, err
	}
	if item.Type != "movie" && item.Type != "tv" {
		return library.ScanJob{}, ErrUnsupported
	}
	job, err := m.db.NewScan(ctx, id)
	if err != nil {
		return job, err
	}
	m.active[id] = true
	m.workers.Add(1)
	go func() {
		defer m.workers.Done()
		defer func() { m.mu.Lock(); delete(m.active, id); m.mu.Unlock() }()
		m.run(item, job)
	}()
	return job, nil
}

// WithIdle serializes configuration changes against scan scheduling.
func (m *Manager) WithIdle(id int64, action func() error) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.closed {
		return ErrClosed
	}
	if m.active[id] {
		return ErrRunning
	}
	return action()
}
func (m *Manager) Close() {
	m.mu.Lock()
	m.closed = true
	m.cancel()
	m.mu.Unlock()
	m.workers.Wait()
}
