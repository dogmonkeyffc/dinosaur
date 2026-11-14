package server

import (
	"database/sql"
	"errors"
	"net/http"
	"strconv"

	"dinosaur/internal/television"
)

func (s *Server) mountWorks(mux *http.ServeMux) {
	mux.Handle("GET /api/works", s.require(false, http.HandlerFunc(s.listWorks)))
	mux.Handle("GET /api/works/{id}", s.require(false, http.HandlerFunc(s.workDetail)))
	mux.Handle("POST /api/admin/works", s.require(true, http.HandlerFunc(s.saveWork)))
	mux.Handle("PUT /api/admin/works/{id}", s.require(true, http.HandlerFunc(s.saveWork)))
	mux.Handle("DELETE /api/admin/works/{id}", s.require(true, http.HandlerFunc(s.deleteWork)))
}

func setWorkPosters(detail *television.WorkDetail) {
	for i := range detail.Members {
		setSeriesPoster(&detail.Members[i])
	}
	for i := range detail.AllMembers {
		setSeriesPoster(&detail.AllMembers[i])
	}
	if len(detail.Members) > 0 {
		detail.Work.PosterURL = detail.Members[0].PosterURL
	}
}

func (s *Server) listWorks(w http.ResponseWriter, r *http.Request) {
	var libraryID int64
	if raw, present := r.URL.Query()["library_id"]; present {
		var err error
		if len(raw) != 1 {
			fail(w, 400, "invalid_library_id", "invalid library ID")
			return
		}
		libraryID, err = strconv.ParseInt(raw[0], 10, 64)
		if err != nil || libraryID <= 0 {
			fail(w, 400, "invalid_library_id", "invalid library ID")
			return
		}
	}
	items, err := s.store.Works(r.Context(), libraryID, current(r).user.ID)
	if s.workError(w, r, err) {
		return
	}
	for i := range items {
		detail, err := s.store.WorkByID(r.Context(), items[i].ID, current(r).user.ID)
		if s.workError(w, r, err) {
			return
		}
		setWorkPosters(&detail)
		items[i] = detail.Work
	}
	writeJSON(w, 200, map[string]any{"items": items})
}

func (s *Server) workDetail(w http.ResponseWriter, r *http.Request) {
	id, ok := libraryID(w, r)
	if !ok {
		return
	}
	detail, err := s.store.WorkByID(r.Context(), id, current(r).user.ID)
	if s.workError(w, r, err) {
		return
	}
	setWorkPosters(&detail)
	if current(r).user.Role != "admin" {
		detail.AllMembers = nil
	}
	writeJSON(w, 200, detail)
}

func (s *Server) saveWork(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Title     string  `json:"title"`
		SeriesIDs []int64 `json:"series_ids"`
	}
	if !decodeJSON(w, r, &body) {
		return
	}
	var id int64
	if r.Method == http.MethodPut {
		var ok bool
		id, ok = libraryID(w, r)
		if !ok {
			return
		}
	}
	work, err := s.store.SaveWork(r.Context(), id, body.Title, body.SeriesIDs)
	if s.workError(w, r, err) {
		return
	}
	detail, err := s.store.WorkByID(r.Context(), work.ID, current(r).user.ID)
	if s.workError(w, r, err) {
		return
	}
	setWorkPosters(&detail)
	status := http.StatusOK
	if r.Method == http.MethodPost {
		status = http.StatusCreated
	}
	writeJSON(w, status, map[string]any{"work": detail.Work})
}

func (s *Server) deleteWork(w http.ResponseWriter, r *http.Request) {
	id, ok := libraryID(w, r)
	if !ok {
		return
	}
	if s.workError(w, r, s.store.DeleteWork(r.Context(), id)) {
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) workError(w http.ResponseWriter, r *http.Request, err error) bool {
	if err == nil {
		return false
	}
	var invalid *television.WorkError
	switch {
	case errors.As(err, &invalid):
		status := http.StatusBadRequest
		if invalid.Code == "work_conflict" {
			status = http.StatusConflict
		}
		fail(w, status, invalid.Code, invalid.Message)
	case errors.Is(err, sql.ErrNoRows):
		fail(w, 404, "not_found", "work not found")
	default:
		s.internal(w, r, err)
	}
	return true
}
