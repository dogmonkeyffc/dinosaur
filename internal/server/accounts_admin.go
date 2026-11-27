package server

import (
	"database/sql"
	"errors"
	"net/http"
	"strconv"

	"dinosaur/internal/auth"
	"dinosaur/internal/store"
)

func (s *Server) mountAccountAdministration(mux *http.ServeMux) {
	mux.Handle("PUT /api/admin/users/{id}/password", s.require(true, http.HandlerFunc(s.resetUserPassword)))
	mux.Handle("GET /api/admin/users/{id}/libraries", s.require(true, http.HandlerFunc(s.userLibraries)))
	mux.Handle("PATCH /api/admin/users/{id}/libraries", s.require(true, http.HandlerFunc(s.updateUserLibraries)))
}

func accountID(w http.ResponseWriter, r *http.Request) (int64, bool) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		fail(w, 400, "invalid_id", "invalid user ID")
		return 0, false
	}
	return id, true
}

func (s *Server) resetUserPassword(w http.ResponseWriter, r *http.Request) {
	id, ok := accountID(w, r)
	if !ok {
		return
	}
	actor := current(r).user
	if id == actor.ID {
		fail(w, 409, "use_account_password", "change your own password in account settings")
		return
	}
	if !s.rate(w, auth.Limit{Key: "reset-password:" + strconv.FormatInt(actor.ID, 10), Max: 20}) {
		return
	}
	var body struct {
		Password string `json:"new_password"`
	}
	if !decodeJSON(w, r, &body) {
		return
	}
	ctx := r.Context()
	target, err := s.store.UserByID(ctx, id)
	if s.accountAdministrationError(w, r, err) {
		return
	}
	if err := auth.ValidateCredentials(target.Username, body.Password); err != nil {
		fail(w, 400, "invalid_password", err.Error())
		return
	}
	hash, err := s.hasher.Hash(ctx, body.Password)
	if err != nil {
		s.internal(w, r, err)
		return
	}
	if !s.accountAdministrationError(w, r, s.store.ChangePassword(ctx, id, hash)) {
		w.WriteHeader(http.StatusNoContent)
	}
}

func (s *Server) userLibraries(w http.ResponseWriter, r *http.Request) {
	id, ok := accountID(w, r)
	if !ok {
		return
	}
	rows, err := s.store.UserLibraries(r.Context(), id)
	if !s.accountAdministrationError(w, r, err) {
		w.Header().Set("Cache-Control", "no-store")
		writeJSON(w, 200, map[string]any{"libraries": rows})
	}
}

func (s *Server) updateUserLibraries(w http.ResponseWriter, r *http.Request) {
	id, ok := accountID(w, r)
	if !ok {
		return
	}
	var body struct {
		Changes []struct {
			LibraryID int64 `json:"library_id"`
			Granted   *bool `json:"granted"`
		} `json:"changes"`
	}
	if !decodeJSON(w, r, &body) {
		return
	}
	changes := make(map[int64]bool, len(body.Changes))
	for _, change := range body.Changes {
		if _, duplicate := changes[change.LibraryID]; change.LibraryID <= 0 || change.Granted == nil || duplicate {
			fail(w, 400, "invalid_library_access", "provide a grant decision per library")
			return
		}
		changes[change.LibraryID] = *change.Granted
	}
	if len(changes) == 0 {
		fail(w, 400, "invalid_library_access", "provide a grant decision per library")
		return
	}
	if !s.accountAdministrationError(w, r, s.store.ChangeUserLibraries(r.Context(), id, changes)) {
		w.WriteHeader(http.StatusNoContent)
	}
}

func (s *Server) accountAdministrationError(w http.ResponseWriter, r *http.Request, err error) bool {
	switch {
	case err == nil:
		return false
	case errors.Is(err, sql.ErrNoRows):
		fail(w, 404, "not_found", "user or media library not found")
	case errors.Is(err, store.ErrLibraryAccount):
		fail(w, 409, "administrator_library_access", "administrators can access all libraries")
	case errors.Is(err, store.ErrLibraryUnrestricted):
		fail(w, 409, "library_unrestricted", "public media libraries do not use per-user grants")
	default:
		s.internal(w, r, err)
	}
	return true
}
