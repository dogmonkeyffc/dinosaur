package server

import (
	"net/http"

	"dinosaur/internal/playmetrics"
)

func (s *Server) mountPlaybackMetrics(mux *http.ServeMux) {
	mux.Handle("POST /api/playback/metrics", s.require(false, http.HandlerFunc(s.savePlaybackMetrics)))
	mux.Handle("GET /api/admin/playback/metrics", s.require(true, http.HandlerFunc(s.listPlaybackMetrics)))
}

// Pointers distinguish required zero-valued measurements from missing fields.
// The shared payload never includes stream_id; it is decoded only on writes.
type playbackMetricPayload struct {
	ID            string `json:"id"`
	Mode          string `json:"mode"`
	Quality       string `json:"quality"`
	BurnSubtitle  *bool  `json:"burn_subtitle,omitempty"`
	Width         *int64 `json:"width"`
	Height        *int64 `json:"height"`
	StartupMS     *int64 `json:"startup_ms,omitempty"`
	Stalls        *int64 `json:"stalls"`
	StallMS       *int64 `json:"stall_ms"`
	PlayedMS      *int64 `json:"played_ms"`
	BufferMS      *int64 `json:"buffer_ms"`
	DroppedFrames *int64 `json:"dropped_frames,omitempty"`
	TotalFrames   *int64 `json:"total_frames,omitempty"`
}

type playbackMetricWrite struct {
	playbackMetricPayload
	StreamID               *string `json:"stream_id"`
	MediaID                *int64  `json:"media_id"`
	AudioIndex             *int    `json:"audio_index"`
	CompatibilityRequested bool    `json:"compatibility_requested"`
}

type playbackMetricResponse struct {
	playbackMetricPayload
	UpdatedAt      int64               `json:"updated_at"`
	TranscodeSpeed *float64            `json:"transcode_speed,omitempty"`
	Source         *playmetrics.Source `json:"source,omitempty"`
	Reason         string              `json:"reason,omitempty"`
}

func (s *Server) savePlaybackMetrics(w http.ResponseWriter, r *http.Request) {
	id := current(r)
	if id.device {
		fail(w, http.StatusForbidden, "forbidden", "browser session required")
		return
	}
	var p playbackMetricWrite
	if !decodeJSON(w, r, &p) {
		return
	}
	if p.Width == nil || p.Height == nil || p.Stalls == nil || p.StallMS == nil || p.PlayedMS == nil || p.BufferMS == nil {
		fail(w, http.StatusBadRequest, "invalid_playback_metrics", "width, height, stalls, stall_ms, played_ms and buffer_ms are required")
		return
	}
	snapshot := playmetrics.Snapshot{
		StreamID: p.StreamID,
		Metrics: playmetrics.Metrics{
			ID: p.ID, Mode: p.Mode, Quality: p.Quality, BurnSubtitle: p.BurnSubtitle, Width: *p.Width, Height: *p.Height,
			StartupMS: p.StartupMS, Stalls: *p.Stalls, StallMS: *p.StallMS, PlayedMS: *p.PlayedMS,
			BufferMS: *p.BufferMS, DroppedFrames: p.DroppedFrames, TotalFrames: p.TotalFrames,
		},
	}
	if err := snapshot.Validate(); err != nil {
		fail(w, http.StatusBadRequest, "invalid_playback_metrics", err.Error())
		return
	}
	if (p.MediaID != nil && *p.MediaID <= 0) || (p.AudioIndex != nil && *p.AudioIndex < 0) {
		fail(w, http.StatusBadRequest, "invalid_playback_metrics", "invalid media or audio index")
		return
	}
	if p.MediaID != nil {
		snapshot.Source = s.playbackMetricSource(r, p)
		snapshot.Reason = playbackMetricReason(p)
	}
	if err := s.store.SavePlaybackMetrics(r.Context(), id.user.ID, snapshot); err != nil {
		s.internal(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) listPlaybackMetrics(w http.ResponseWriter, r *http.Request) {
	items, err := s.store.ListPlaybackMetrics(r.Context())
	if err != nil {
		s.internal(w, r, err)
		return
	}
	response := make([]playbackMetricResponse, 0, len(items))
	for _, item := range items {
		response = append(response, playbackMetricResponse{
			playbackMetricPayload: playbackMetricPayload{
				ID: item.ID, Mode: item.Mode, Quality: item.Quality, BurnSubtitle: item.BurnSubtitle, Width: &item.Width, Height: &item.Height,
				StartupMS: item.StartupMS, Stalls: &item.Stalls, StallMS: &item.StallMS, PlayedMS: &item.PlayedMS,
				BufferMS: &item.BufferMS, DroppedFrames: item.DroppedFrames, TotalFrames: item.TotalFrames,
			},
			UpdatedAt: item.UpdatedAt, TranscodeSpeed: item.TranscodeSpeed, Source: item.Source, Reason: item.Reason,
		})
	}
	writeJSON(w, http.StatusOK, struct {
		Items []playbackMetricResponse `json:"items"`
	}{Items: response})
}
