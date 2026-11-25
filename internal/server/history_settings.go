package server

import "net/http"

func (s *Server) listHistorySettings(w http.ResponseWriter, r *http.Request) {
	items, err := s.store.HistoryLibrarySettings(r.Context(), current(r).user.ID)
	if err != nil {
		s.historyError(w, r, err)
		return
	}
	writeJSON(w, 200, map[string]any{"libraries": items})
}

func (s *Server) updateHistorySetting(w http.ResponseWriter, r *http.Request) {
	id, ok := libraryID(w, r)
	if !ok {
		return
	}
	var body struct {
		Record *bool `json:"record_history"`
	}
	if !decodeJSON(w, r, &body) {
		return
	}
	if body.Record == nil {
		fail(w, 400, "invalid_history_setting", "record_history is required")
		return
	}
	if err := s.store.SetHistoryLibrarySetting(r.Context(), current(r).user.ID, id, *body.Record); err != nil {
		s.historyError(w, r, err)
		return
	}
	writeJSON(w, 200, map[string]any{"library_id": id, "record_history": *body.Record})
}
