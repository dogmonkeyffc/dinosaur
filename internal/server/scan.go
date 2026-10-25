package server

import (
	"context"
	"database/sql"
	"errors"
	"net/http"

	"dinosaur/internal/scanner"
)

func (s *Server) startScan(w http.ResponseWriter, r *http.Request) {
	id, ok := libraryID(w, r)
	if !ok {
		return
	}
	job, err := s.scanner.Scan(r.Context(), id)
	if s.scanError(w, r, err) {
		return
	}
	writeJSON(w, 202, map[string]any{"scan": job})
}
func (s *Server) scanStatus(w http.ResponseWriter, r *http.Request) {
	id, ok := libraryID(w, r)
	if !ok {
		return
	}
	if _, err := s.store.LibraryByID(r.Context(), id); s.scanError(w, r, err) {
		return
	}
	job, err := s.store.LatestScan(r.Context(), id)
	if err != nil {
		s.internal(w, r, err)
		return
	}
	writeJSON(w, 200, map[string]any{"scan": job})
}
func (s *Server) scanError(w http.ResponseWriter, r *http.Request, err error) bool {
	if err == nil {
		return false
	}
	switch {
	case errors.Is(err, sql.ErrNoRows):
		fail(w, 404, "not_found", "library not found")
	case errors.Is(err, scanner.ErrRunning):
		fail(w, 409, "scan_running", "library scan is active; retry after it finishes")
	case errors.Is(err, scanner.ErrUnsupported):
		fail(w, 400, "unsupported_library", "unsupported library type")
	case errors.Is(err, scanner.ErrClosed):
		fail(w, 503, "scanner_unavailable", "scanner is shutting down")
	default:
		s.internal(w, r, err)
	}
	return true
}

func (s *Server) StartScanning() error {
	if err := s.store.InterruptScrapes(context.Background()); err != nil {
		return err
	}
	return s.scanner.Start()
}
func (s *Server) Close() {
	s.scanner.Close()
	s.scraper.Close()
	s.tvScraper.Close()
	s.transcoder.Close()
}
