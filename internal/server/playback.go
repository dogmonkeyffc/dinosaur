package server

import (
	"database/sql"
	"errors"
	"net/http"
	"strconv"
	"time"

	"dinosaur/internal/media"
	"dinosaur/internal/playback"
	"dinosaur/internal/streaming"
)

func (s *Server) mountPlayback(mux *http.ServeMux) {
	mux.Handle("GET /api/playback/{id}", s.require(false, http.HandlerFunc(s.playbackInfo)))
	mux.Handle("PUT /api/playback/{id}/progress", s.require(false, http.HandlerFunc(s.savePlaybackProgress)))
	mux.Handle("GET /api/playback/continue", s.require(false, http.HandlerFunc(s.continueWatching)))
	mux.Handle("GET /api/playback/{id}/subtitles/{subtitle}", s.require(false, http.HandlerFunc(s.playbackSubtitle)))
}
func (s *Server) playbackInfo(w http.ResponseWriter, r *http.Request) {
	source, ok := s.playbackSource(w, r)
	if !ok {
		return
	}
	plan, err := requestPlan(r, source.Item)
	if err != nil {
		fail(w, 400, "invalid_playback", err.Error())
		return
	}
	progress, err := s.store.PlaybackProgress(r.Context(), current(r).user.ID, source.Item.ID)
	if err != nil {
		s.internal(w, r, err)
		return
	}
	progress.Duration = source.Item.Duration
	audio := make([]media.Track, 0)
	for _, track := range source.Item.Tracks {
		if track.Type == "audio" {
			audio = append(audio, track)
		}
	}
	setPosterURL(&source.Item)
	writeJSON(w, 200, map[string]any{
		"item": source.Item, "plan": plan, "stream_url": streamURL(source.Item.ID, plan, r.URL.Query().Get("mode") == "compatible"),
		"progress": progress, "resume_position": progress.ResumePosition(), "audio_tracks": audio,
		"subtitles": streaming.Subtitles(source.Root, source.Item.RelativePath),
	})
}
func (s *Server) savePlaybackProgress(w http.ResponseWriter, r *http.Request) {
	id, ok := libraryID(w, r)
	if !ok {
		return
	}
	var body struct {
		Position *float64 `json:"position"`
		Watched  *bool    `json:"watched"`
	}
	if !decodeJSON(w, r, &body) {
		return
	}
	if body.Position == nil && body.Watched == nil {
		fail(w, 400, "invalid_progress", "position or watched is required")
		return
	}
	item, err := s.store.MediaByID(r.Context(), id)
	if err != nil {
		s.playbackError(w, err)
		return
	}
	progress, err := s.store.PlaybackProgress(r.Context(), current(r).user.ID, id)
	if err != nil {
		s.internal(w, r, err)
		return
	}
	progress, err = playback.UpdateProgress(progress, item.Duration, body.Position, body.Watched)
	if err != nil {
		fail(w, 400, "invalid_progress", err.Error())
		return
	}
	progress.UpdatedAt = time.Now().Unix()
	if err := s.store.SavePlaybackProgress(r.Context(), current(r).user.ID, progress); err != nil {
		s.internal(w, r, err)
		return
	}
	writeJSON(w, 200, map[string]any{"progress": progress})
}
func (s *Server) continueWatching(w http.ResponseWriter, r *http.Request) {
	items, err := s.store.ContinueWatching(r.Context(), current(r).user.ID)
	if err != nil {
		s.internal(w, r, err)
		return
	}
	for i := range items {
		setPosterURL(&items[i])
	}
	writeJSON(w, 200, map[string]any{"items": items})
}
func (s *Server) playbackSubtitle(w http.ResponseWriter, r *http.Request) {
	source, ok := s.playbackSource(w, r)
	if !ok {
		return
	}
	id, err := strconv.Atoi(r.PathValue("subtitle"))
	subtitles := streaming.Subtitles(source.Root, source.Item.RelativePath)
	if err != nil || id < 0 || id >= len(subtitles) {
		fail(w, 404, "not_found", "subtitle not found")
		return
	}
	data, err := streaming.ReadSubtitle(source.Root, subtitles[id].RelativePath)
	if err != nil {
		fail(w, 422, "subtitle_unavailable", "subtitle must be UTF-8 WebVTT or SRT under 2 MiB")
		return
	}
	w.Header().Set("Content-Type", "text/vtt; charset=utf-8")
	_, _ = w.Write(data)
}
func (s *Server) playbackError(w http.ResponseWriter, err error) {
	if errors.Is(err, sql.ErrNoRows) {
		fail(w, 404, "not_found", "media not found")
		return
	}
	s.logger.Error("playback source lookup failed", "error", err)
	fail(w, 500, "internal_error", "unable to read media information")
}
