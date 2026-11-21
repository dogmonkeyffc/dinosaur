package server

import (
	"errors"
	"net/http"

	"dinosaur/internal/agent"
	"dinosaur/internal/scanner"
	"dinosaur/internal/television"
)

func (s *Server) agentImportEpisodes(w http.ResponseWriter, r *http.Request) {
	id, ok := libraryID(w, r)
	if !ok {
		return
	}
	if !agentKey(r).Allows(id, true) {
		s.agentError(w, r, agent.ErrForbidden)
		return
	}
	var body struct {
		Episodes []struct {
			Path         string `json:"path"`
			SeriesFolder string `json:"series_folder"`
			Season       *int   `json:"season"`
			Episode      int    `json:"episode"`
		} `json:"episodes"`
	}
	if !decodeJSON(w, r, &body) {
		return
	}
	if len(body.Episodes) == 0 || len(body.Episodes) > 100 {
		fail(w, 400, "invalid_episode_assignment", "send 1 to 100 episode assignments")
		return
	}
	assignments := make([]television.EpisodeAssignment, 0, len(body.Episodes))
	for _, episode := range body.Episodes {
		if episode.Season == nil {
			fail(w, 400, "invalid_episode_assignment", "season is required")
			return
		}
		assignments = append(assignments, television.EpisodeAssignment{
			Path: episode.Path, SeriesFolder: episode.SeriesFolder, Season: *episode.Season, Episode: episode.Episode,
		})
	}
	job, err := s.scanner.ImportEpisodes(r.Context(), id, assignments)
	switch {
	case errors.Is(err, scanner.ErrRunning):
		fail(w, 409, "scan_running", "wait for the current scan before importing episodes")
	case errors.Is(err, television.ErrEpisodeAssigned):
		fail(w, 409, "episode_already_assigned", err.Error())
	case errors.Is(err, television.ErrEpisodeAssignment), errors.Is(err, scanner.ErrUnsupported):
		fail(w, 400, "invalid_episode_assignment", "send valid video paths, series folders and episode numbers for a television library")
	default:
		if !s.agentError(w, r, err) {
			writeJSON(w, 202, map[string]any{"scan": job})
		}
	}
}

func (s *Server) agentScanStatus(w http.ResponseWriter, r *http.Request) {
	id, ok := libraryID(w, r)
	if !ok {
		return
	}
	if !agentKey(r).Allows(id, false) {
		s.agentError(w, r, agent.ErrForbidden)
		return
	}
	if _, err := s.store.LibraryByID(r.Context(), id); s.agentError(w, r, err) {
		return
	}
	scan, err := s.store.LatestScan(r.Context(), id)
	if !s.agentError(w, r, err) {
		w.Header().Set("Cache-Control", "no-store")
		writeJSON(w, 200, map[string]any{"scan": scan})
	}
}
