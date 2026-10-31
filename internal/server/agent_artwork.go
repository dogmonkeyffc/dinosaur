package server

import (
	"errors"
	"net/http"
	"strconv"

	"dinosaur/internal/artwork"
	"dinosaur/internal/metadata"
)

func (s *Server) putAgentArtwork(w http.ResponseWriter, r *http.Request) {
	target, ok := s.scopedAgentTarget(w, r, true)
	if !ok {
		return
	}
	expected, ok := agentVersion(w, r)
	if !ok {
		return
	}
	if expected != target.Version {
		fail(w, 409, "metadata_conflict", "metadata changed; read its current version")
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, artwork.MaxSize+1)
	key, err := artwork.Cache(r.Context(), r.Body, s.config.DataDir)
	if errors.Is(err, artwork.ErrTooLarge) {
		fail(w, 413, "body_too_large", "artwork exceeds 5 MiB")
		return
	}
	if errors.Is(err, artwork.ErrImage) {
		fail(w, 400, "invalid_artwork", "upload JPEG or PNG up to 5 MiB and 8 megapixels")
		return
	}
	if s.agentError(w, r, err) {
		return
	}
	version, err := s.agentService().Artwork(r.Context(), agentKey(r), target, expected, key)
	if errors.Is(err, metadata.ErrConflict) {
		w.Header().Set("ETag", `"`+strconv.FormatInt(version, 10)+`"`)
	}
	if s.agentError(w, r, err) {
		return
	}
	w.Header().Set("ETag", `"`+strconv.FormatInt(version, 10)+`"`)
	writeJSON(w, 200, map[string]any{"version": version, "confirmed": true})
}

func (s *Server) getAgentArtwork(w http.ResponseWriter, r *http.Request) {
	target, ok := s.scopedAgentTarget(w, r, false)
	if !ok {
		return
	}
	if target.Metadata == nil || target.Metadata.PosterKey == "" {
		fail(w, 404, "not_found", "no metadata artwork")
		return
	}
	s.servePoster(w, r, target.Metadata.PosterKey)
}
