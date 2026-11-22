package server

import (
	"database/sql"
	"errors"
	"net/http"

	"dinosaur/internal/store"
)

func (s *Server) mountLibraryAccess(mux *http.ServeMux) {
	mux.Handle("GET /api/admin/libraries/{id}/access", s.require(true, http.HandlerFunc(s.libraryAccess)))
	mux.Handle("PUT /api/admin/libraries/{id}/access", s.require(true, http.HandlerFunc(s.saveLibraryAccess)))
}

func (s *Server) libraryAccess(w http.ResponseWriter, r *http.Request) {
	id, ok := libraryID(w, r)
	if !ok {
		return
	}
	access, accounts, err := s.store.LibraryAccess(r.Context(), id)
	if s.libraryAccessError(w, r, err) {
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	writeJSON(w, 200, map[string]any{"access": access, "accounts": accounts})
}

func (s *Server) saveLibraryAccess(w http.ResponseWriter, r *http.Request) {
	id, ok := libraryID(w, r)
	if !ok {
		return
	}
	var body struct {
		Restricted *bool   `json:"restricted"`
		UserIDs    []int64 `json:"user_ids"`
	}
	if !decodeJSON(w, r, &body) {
		return
	}
	seen := map[int64]bool{}
	for _, id := range body.UserIDs {
		if id < 1 || seen[id] {
			s.libraryAccessError(w, r, store.ErrLibraryAccount)
			return
		}
		seen[id] = true
	}
	if body.Restricted == nil {
		s.libraryAccessError(w, r, store.ErrLibraryAccount)
		return
	}
	access := store.LibraryAccess{Restricted: *body.Restricted, UserIDs: body.UserIDs}
	if !s.libraryAccessError(w, r, s.store.SaveLibraryAccess(r.Context(), id, access)) {
		w.WriteHeader(http.StatusNoContent)
	}
}

func (s *Server) libraryAccessError(w http.ResponseWriter, r *http.Request, err error) bool {
	switch {
	case err == nil:
		return false
	case errors.Is(err, sql.ErrNoRows):
		fail(w, 404, "not_found", "library not found")
	case errors.Is(err, store.ErrLibraryAccount):
		fail(w, 400, "invalid_library_access", "select valid accounts")
	default:
		s.internal(w, r, err)
	}
	return true
}
