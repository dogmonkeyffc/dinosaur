package server

import (
	"errors"
	"net/http"
	"strconv"

	"dinosaur/internal/auth"
	"dinosaur/internal/metadata"
	"dinosaur/internal/tmdb"
)

func (s *Server) mountMetadataSettings(mux *http.ServeMux) {
	mux.Handle("GET /api/admin/metadata/tmdb", s.require(true, http.HandlerFunc(s.getMetadataSettings)))
	mux.Handle("PUT /api/admin/metadata/tmdb", s.require(true, http.HandlerFunc(s.saveMetadataSettings)))
	mux.Handle("DELETE /api/admin/metadata/tmdb", s.require(true, http.HandlerFunc(s.clearMetadataCredential)))
	mux.Handle("POST /api/admin/metadata/tmdb/test", s.require(true, http.HandlerFunc(s.testMetadataConnection)))
}

func writeMetadataSettings(w http.ResponseWriter, settings metadata.Settings) {
	writeJSON(w, 200, map[string]any{
		"credential_type": settings.CredentialType,
		"language":        settings.Language,
		"configured":      settings.Credential != "",
	})
}

func (s *Server) getMetadataSettings(w http.ResponseWriter, r *http.Request) {
	settings, err := s.store.MetadataSettings(r.Context())
	if err != nil {
		s.internal(w, r, err)
		return
	}
	writeMetadataSettings(w, settings)
}

func (s *Server) saveMetadataSettings(w http.ResponseWriter, r *http.Request) {
	var update metadata.SettingsUpdate
	if !decodeJSON(w, r, &update) {
		return
	}
	settings, err := s.store.UpdateMetadataSettings(r.Context(), update, false)
	if s.metadataSettingsError(w, r, err) {
		return
	}
	writeMetadataSettings(w, settings)
}

func (s *Server) clearMetadataCredential(w http.ResponseWriter, r *http.Request) {
	settings, err := s.store.UpdateMetadataSettings(r.Context(), metadata.SettingsUpdate{}, true)
	if s.metadataSettingsError(w, r, err) {
		return
	}
	writeMetadataSettings(w, settings)
}

func (s *Server) metadataSettingsError(w http.ResponseWriter, r *http.Request, err error) bool {
	if errors.Is(err, metadata.ErrSettings) {
		fail(w, 400, "invalid_metadata_settings", "invalid credential type, language or credential")
	} else if err != nil {
		s.internal(w, r, err)
	}
	return err != nil
}

func (s *Server) testMetadataConnection(w http.ResponseWriter, r *http.Request) {
	if !s.rate(w, auth.Limit{Key: "tmdb-test:" + strconv.FormatInt(current(r).user.ID, 10), Max: 20}) {
		return
	}
	var update metadata.SettingsUpdate
	if !decodeJSON(w, r, &update) {
		return
	}
	settings, err := s.store.MetadataSettings(r.Context())
	if s.metadataSettingsError(w, r, err) {
		return
	}
	settings, err = settings.Apply(update)
	if s.metadataSettingsError(w, r, err) {
		return
	}
	err = s.tmdb.Test(r.Context(), settings)
	switch {
	case err == nil:
		writeJSON(w, 200, map[string]bool{"ok": true})
	case errors.Is(err, metadata.ErrMissingCredential):
		fail(w, 400, "tmdb_not_configured", "TMDB credential is not configured")
	case errors.Is(err, tmdb.ErrUnauthorized):
		fail(w, 422, "tmdb_invalid_credential", "TMDB rejected the credential")
	case errors.Is(err, tmdb.ErrRateLimited):
		fail(w, 503, "tmdb_rate_limited", "TMDB rate limit exceeded")
	default:
		fail(w, 502, "tmdb_unavailable", "cannot connect to TMDB")
	}
}
