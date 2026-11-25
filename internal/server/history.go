package server

import (
	"database/sql"
	"errors"
	"net/http"
	"path/filepath"
	"strconv"

	"dinosaur/internal/auth"
	"dinosaur/internal/history"
	"dinosaur/internal/historyassets"
)

func (s *Server) mountHistory(mux *http.ServeMux) {
	mux.Handle("GET /api/history/settings", s.require(false, http.HandlerFunc(s.listHistorySettings)))
	mux.Handle("PUT /api/history/settings/{id}", s.require(false, http.HandlerFunc(s.updateHistorySetting)))
	mux.Handle("GET /api/history", s.require(false, http.HandlerFunc(s.listHistory)))
	mux.Handle("POST /api/history", s.require(false, http.HandlerFunc(s.createHistory)))
	mux.Handle("PUT /api/history/{id}", s.require(false, http.HandlerFunc(s.updateHistory)))
	mux.Handle("DELETE /api/history/{id}", s.require(false, http.HandlerFunc(s.deleteHistory)))
	mux.Handle("GET /api/history/{id}/poster", s.require(false, http.HandlerFunc(s.historyPoster)))
}

func historyURLs(e *history.Entry) {
	if history.ValidPosterKey(e.PosterKey) || (e.Source == "local" && e.MediaID != nil) {
		e.PosterURL = "/api/history/" + strconv.FormatInt(e.ID, 10) + "/poster"
	}
	if e.Source == "douban" {
		e.SourceURL = "https://movie.douban.com/subject/" + e.SourceID + "/"
	}
}

func (s *Server) listHistory(w http.ResponseWriter, r *http.Request) {
	page, size := 1, 24
	for key, dst := range map[string]*int{"page": &page, "page_size": &size} {
		if raw := r.URL.Query().Get(key); raw != "" {
			value, err := strconv.Atoi(raw)
			if err != nil {
				fail(w, 400, "invalid_filter", "invalid pagination")
				return
			}
			*dst = value
		}
	}
	result, err := s.store.HistoryPage(r.Context(), current(r).user.ID, page, size)
	if err != nil {
		s.historyError(w, r, err)
		return
	}
	for i := range result.Items {
		historyURLs(&result.Items[i])
	}
	writeJSON(w, 200, result)
}

func (s *Server) createHistory(w http.ResponseWriter, r *http.Request) {
	var in history.Input
	if !decodeJSON(w, r, &in) {
		return
	}
	sourceID, err := auth.NewToken()
	if err != nil {
		s.internal(w, r, err)
		return
	}
	e := history.Entry{Source: "manual", SourceID: sourceID}
	e.Apply(in)
	e.ID, err = s.store.CreateHistory(r.Context(), current(r).user.ID, e)
	if err != nil {
		s.historyError(w, r, err)
		return
	}
	writeJSON(w, 201, map[string]any{"item": e})
}

func (s *Server) updateHistory(w http.ResponseWriter, r *http.Request) {
	id, ok := libraryID(w, r)
	if !ok {
		return
	}
	var in history.Input
	if !decodeJSON(w, r, &in) {
		return
	}
	e, err := s.store.HistoryByID(r.Context(), current(r).user.ID, id)
	if err != nil {
		s.historyError(w, r, err)
		return
	}
	e.Apply(in)
	if err = s.store.UpdateHistory(r.Context(), current(r).user.ID, e); err != nil {
		s.historyError(w, r, err)
		return
	}
	historyURLs(&e)
	writeJSON(w, 200, map[string]any{"item": e})
}

func (s *Server) deleteHistory(w http.ResponseWriter, r *http.Request) {
	id, ok := libraryID(w, r)
	if !ok {
		return
	}
	if err := s.store.DeleteHistory(r.Context(), current(r).user.ID, id); err != nil {
		s.historyError(w, r, err)
		return
	}
	writeJSON(w, 200, map[string]bool{"ok": true})
}

func (s *Server) historyPoster(w http.ResponseWriter, r *http.Request) {
	id, ok := libraryID(w, r)
	if !ok {
		return
	}
	e, err := s.store.HistoryByID(r.Context(), current(r).user.ID, id)
	if err != nil {
		s.historyError(w, r, err)
		return
	}
	if e.Source == "local" && e.MediaID != nil {
		item, err := s.store.MediaByID(r.Context(), *e.MediaID)
		if err != nil {
			fail(w, 404, "not_found", "poster unavailable")
			return
		}
		s.servePoster(w, r, item.PosterKey)
		return
	}
	f, err := historyassets.Open(filepath.Join(s.config.DataDir, "history-posters"), e.PosterKey)
	if err != nil {
		fail(w, 404, "not_found", "poster unavailable")
		return
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil || !info.Mode().IsRegular() {
		fail(w, 404, "not_found", "poster unavailable")
		return
	}
	w.Header().Set("Cache-Control", "private, no-store")
	w.Header().Del("Content-Type")
	http.ServeContent(w, r, e.PosterKey, info.ModTime(), f)
}

func (s *Server) historyError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, sql.ErrNoRows):
		fail(w, 404, "not_found", "watch history record not found")
	case errors.Is(err, history.ErrInvalid):
		fail(w, 400, "invalid_history", "invalid title, date, year or rating")
	case errors.Is(err, history.ErrConflict):
		fail(w, 409, "history_conflict", "watch history record already exists")
	default:
		s.internal(w, r, err)
	}
}
