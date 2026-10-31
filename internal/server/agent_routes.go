package server

import (
	"database/sql"
	"errors"
	"net/http"
	"strconv"

	"dinosaur/internal/agent"
	"dinosaur/internal/metadata"
)

func (s *Server) mountAgent(mux *http.ServeMux) {
	mux.Handle("GET /api/admin/agent-keys/docs", s.require(true, http.HandlerFunc(s.agentDocs)))
	mux.Handle("GET /api/agent/help", s.requireAgent(false, http.HandlerFunc(s.agentDocs)))
	mux.Handle("GET /api/admin/agent-keys", s.require(true, http.HandlerFunc(s.listAgentKeys)))
	mux.Handle("POST /api/admin/agent-keys", s.require(true, http.HandlerFunc(s.createAgentKey)))
	mux.Handle("DELETE /api/admin/agent-keys/{id}", s.require(true, http.HandlerFunc(s.revokeAgentKey)))
	mux.Handle("GET /api/agent/tasks", s.requireAgent(false, http.HandlerFunc(s.agentTasks)))
	mux.Handle("GET /api/agent/{kind}/{id}", s.requireAgent(false, http.HandlerFunc(s.agentTarget)))
	mux.Handle("GET /api/agent/{kind}/{id}/metadata", s.requireAgent(false, http.HandlerFunc(s.agentMetadata)))
	mux.Handle("GET /api/agent/{kind}/{id}/metadata.nfo", s.requireAgent(false, http.HandlerFunc(s.agentNFO)))
	mux.Handle("PUT /api/agent/{kind}/{id}/metadata", s.requireAgent(true, http.HandlerFunc(s.putAgentMetadata)))
	mux.Handle("POST /api/agent/{kind}/{id}/artwork", s.requireAgent(true, http.HandlerFunc(s.putAgentArtwork)))
	mux.Handle("GET /api/agent/{kind}/{id}/artwork", s.requireAgent(false, http.HandlerFunc(s.getAgentArtwork)))
	mux.Handle("/api/agent/", s.requireAgent(false, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { fail(w, 404, "not_found", "unknown agent endpoint") })))
}

func (s *Server) agentService() agent.Service { return agent.Service{Repository: s.store} }

func (s *Server) agentError(w http.ResponseWriter, r *http.Request, err error) bool {
	switch {
	case err == nil:
		return false
	case errors.Is(err, sql.ErrNoRows):
		fail(w, 404, "not_found", "metadata target or key not found")
	case errors.Is(err, agent.ErrMetadataRequired):
		fail(w, 409, "metadata_required", "confirm metadata before uploading artwork")
	case errors.Is(err, agent.ErrInput):
		fail(w, 400, "invalid_agent_input", "invalid metadata, key settings or NFO")
	case errors.Is(err, agent.ErrForbidden):
		fail(w, 403, "agent_forbidden", "key does not permit this library or operation")
	case errors.Is(err, metadata.ErrConflict):
		fail(w, 409, "metadata_conflict", "metadata changed; read the current version before retrying")
	default:
		s.internal(w, r, err)
	}
	return true
}

func (s *Server) scopedAgentTarget(w http.ResponseWriter, r *http.Request, write bool) (agent.Target, bool) {
	kind := "movie"
	switch r.PathValue("kind") {
	case "movies":
	case "series":
		kind = "tv"
	default:
		fail(w, 404, "not_found", "unknown target kind")
		return agent.Target{}, false
	}
	id, ok := libraryID(w, r)
	if !ok {
		return agent.Target{}, false
	}
	target, err := s.agentService().Target(r.Context(), agentKey(r), kind, id, write)
	if s.agentError(w, r, err) {
		return target, false
	}
	w.Header().Set("ETag", `"`+strconv.FormatInt(target.Version, 10)+`"`)
	return target, true
}

func agentVersion(w http.ResponseWriter, r *http.Request) (int64, bool) {
	value := r.Header.Get("If-Match")
	if value == "" {
		fail(w, 428, "version_required", "send the target ETag in If-Match")
		return 0, false
	}
	if len(r.Header.Values("If-Match")) != 1 || len(value) < 3 || value[0] != '"' || value[len(value)-1] != '"' {
		fail(w, 400, "invalid_version", "If-Match must be a quoted metadata version")
		return 0, false
	}
	version, err := strconv.ParseInt(value[1:len(value)-1], 10, 64)
	if err != nil || version < 0 {
		fail(w, 400, "invalid_version", "invalid metadata version")
		return 0, false
	}
	return version, true
}
