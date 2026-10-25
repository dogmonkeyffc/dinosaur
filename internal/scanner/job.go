package scanner

import (
	"context"
	"errors"
	"io/fs"
	"path/filepath"
	"strings"
	"time"

	"dinosaur/internal/library"
	"dinosaur/internal/media"
)

func (m *Manager) run(item library.Library, job library.ScanJob) {
	var err error
	select {
	case m.slots <- struct{}{}:
		defer func() { <-m.slots }()
		err = m.scan(item, &job)
	case <-m.ctx.Done():
		err = m.ctx.Err()
	}
	job.Status = "completed"
	if err != nil {
		job.Status = "failed"
		job.Message = "unable to read library; previous entries were retained"
		if m.ctx.Err() != nil {
			job.Status = "interrupted"
			job.Message = "scan interrupted"
		}
		m.logger.Error("library scan failed", "library_id", item.ID, "error", err)
	}
	if err == nil && m.AfterScan != nil {
		m.AfterScan()
	}
	job.FinishedAt = time.Now().Unix()
	// Job completion must survive cancellation of the worker's context.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if saveErr := m.db.SaveScan(ctx, job); saveErr != nil {
		m.logger.Error("save scan status", "library_id", item.ID, "error", saveErr)
	}
}
func (m *Manager) scan(item library.Library, job *library.ScanJob) error {
	job.Status = "scanning"
	if err := m.db.SaveScan(m.ctx, *job); err != nil {
		return err
	}
	root, err := filepath.EvalSymlinks(item.RootPath)
	if err != nil {
		return err
	}
	if root != item.RootPath || !library.Readable(root) {
		return errors.New("library root is no longer readable")
	}
	paths, err := m.discover(root)
	if err != nil {
		return err
	}
	job.Discovered = len(paths)
	if err := m.db.SaveScan(m.ctx, *job); err != nil {
		return err
	}
	previous, err := m.db.MediaSnapshot(m.ctx, item.ID)
	if err != nil {
		return err
	}
	for _, relative := range paths {
		if err := m.ctx.Err(); err != nil {
			return err
		}
		if err := m.index(item, relative, previous[relative], job); err != nil {
			return err
		}
		job.Processed++
		if err := m.db.SaveScan(m.ctx, *job); err != nil {
			return err
		}
	}
	// A failed traversal or a cancelled scan never marks unseen files missing.
	if err := m.ctx.Err(); err != nil {
		return err
	}
	if !library.Readable(root) {
		return errors.New("library root disconnected during scan")
	}
	return m.db.MarkMissingMedia(m.ctx, item.ID, job.ID)
}
func (m *Manager) discover(root string) ([]string, error) {
	paths := make([]string, 0)
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
		if err := m.ctx.Err(); err != nil {
			return err
		}
		if walkErr != nil {
			return walkErr
		}
		if path != root && strings.HasPrefix(entry.Name(), ".") {
			if entry.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if entry.Type()&fs.ModeSymlink != 0 || entry.IsDir() || !media.IsVideo(path) {
			return nil
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		if !info.Mode().IsRegular() {
			return nil
		}
		relative, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		paths = append(paths, relative)
		return nil
	})
	return paths, err
}
