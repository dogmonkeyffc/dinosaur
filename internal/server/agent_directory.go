package server

import (
	"errors"
	"net/http"

	"dinosaur/internal/agent"
	"dinosaur/internal/media"
)

func (s *Server) agentLibraries(w http.ResponseWriter, r *http.Request) {
	items, err := s.store.Libraries(r.Context())
	if s.agentError(w, r, err) {
		return
	}
	type librarySummary struct {
		ID   int64  `json:"id"`
		Name string `json:"name"`
		Type string `json:"type"`
	}
	libraries := make([]librarySummary, 0, len(items))
	key := agentKey(r)
	for _, item := range items {
		if key.Allows(item.ID, false) {
			libraries = append(libraries, librarySummary{item.ID, item.Name, item.Type})
		}
	}
	w.Header().Set("Cache-Control", "no-store")
	writeJSON(w, http.StatusOK, map[string]any{"libraries": libraries})
}

func (s *Server) agentDirectory(w http.ResponseWriter, r *http.Request) {
	id, ok := libraryID(w, r)
	if !ok {
		return
	}
	if !agentKey(r).Allows(id, false) {
		s.agentError(w, r, agent.ErrForbidden)
		return
	}
	item, err := s.store.LibraryByID(r.Context(), id)
	if s.agentError(w, r, err) {
		return
	}
	directory := r.URL.Query().Get("directory")
	items, next, err := media.DirectoryPage(r.Context(), item.RootPath, directory, r.URL.Query().Get("after"))
	if errors.Is(err, media.ErrInvalidDirectory) {
		fail(w, http.StatusBadRequest, "invalid_directory", "directory must be a relative media path")
		return
	}
	if err != nil {
		fail(w, http.StatusNotFound, "directory_unavailable", "media directory is unavailable")
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	writeJSON(w, http.StatusOK, map[string]any{"library_id": id, "items": items, "next_after": next})
}
