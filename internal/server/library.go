package server

import (
	"database/sql"
	"errors"
	"net/http"
	"strconv"

	"dinosaur/internal/library"
	"dinosaur/internal/scanner"
	"dinosaur/internal/store"
)

type libraryRequest struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	RootPath string `json:"root_path"`
}

func (s *Server) listLibraries(w http.ResponseWriter, r *http.Request) {
	items, err := s.store.Libraries(r.Context())
	if err != nil {
		s.internal(w, r, err)
		return
	}
	result := make([]any, 0, len(items))
	for _, item := range items {
		job, err := s.store.LatestScan(r.Context(), item.ID)
		if err != nil {
			s.internal(w, r, err)
			return
		}
		if current(r).user.Role == "admin" {
			result = append(result, struct {
				store.Library
				Available bool             `json:"available"`
				Scan      *library.ScanJob `json:"scan"`
			}{item, library.Readable(item.RootPath), job})
		} else {
			result = append(result, struct {
				ID    int64  `json:"id"`
				Name  string `json:"name"`
				Type  string `json:"type"`
				Count int    `json:"media_count"`
			}{item.ID, item.Name, item.Type, item.MediaCount})
		}
	}
	writeJSON(w, 200, map[string]any{"libraries": result})
}
func (s *Server) saveLibrary(w http.ResponseWriter, r *http.Request) {
	var body libraryRequest
	if !decodeJSON(w, r, &body) {
		return
	}
	item := store.Library{Name: body.Name, Type: body.Type, RootPath: body.RootPath}
	status := http.StatusCreated
	if r.Method == http.MethodPut {
		id, ok := libraryID(w, r)
		if !ok {
			return
		}
		item.ID = id
		status = http.StatusOK
	}
	item, err := library.Validate(item)
	var invalid *library.ValidationError
	if errors.As(err, &invalid) {
		fail(w, 400, invalid.Code, invalid.Message)
		return
	}
	if err != nil {
		s.internal(w, r, err)
		return
	}
	err = s.scanner.WithIdle(item.ID, func() error {
		var saveErr error
		item, saveErr = s.store.SaveLibrary(r.Context(), item)
		return saveErr
	})
	if errors.Is(err, scanner.ErrRunning) || errors.Is(err, scanner.ErrClosed) {
		s.scanError(w, r, err)
		return
	}
	if errors.Is(err, store.ErrLibraryPathExists) {
		fail(w, 409, "library_path_exists", "this directory is already configured")
		return
	}
	if errors.Is(err, sql.ErrNoRows) {
		fail(w, 404, "not_found", "library not found")
		return
	}
	if err != nil {
		s.internal(w, r, err)
		return
	}
	if item.Type == "movie" {
		if _, err := s.scanner.Scan(r.Context(), item.ID); err != nil {
			s.logger.Warn("schedule library scan", "library_id", item.ID, "error", err)
		}
	}
	writeJSON(w, status, map[string]any{"library": item})
}
func (s *Server) deleteLibrary(w http.ResponseWriter, r *http.Request) {
	id, ok := libraryID(w, r)
	if !ok {
		return
	}
	err := s.scanner.WithIdle(id, func() error { return s.store.DeleteLibrary(r.Context(), id) })
	if errors.Is(err, scanner.ErrRunning) || errors.Is(err, scanner.ErrClosed) {
		s.scanError(w, r, err)
		return
	}
	if errors.Is(err, sql.ErrNoRows) {
		fail(w, 404, "not_found", "library not found")
		return
	}
	if err != nil {
		s.internal(w, r, err)
		return
	}
	writeJSON(w, 200, map[string]bool{"ok": true})
}
func libraryID(w http.ResponseWriter, r *http.Request) (int64, bool) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		fail(w, 400, "invalid_id", "invalid library ID")
		return 0, false
	}
	return id, true
}
