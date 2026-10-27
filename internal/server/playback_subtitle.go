package server

import (
	"errors"
	"math"
	"net/http"
	"strconv"

	"dinosaur/internal/streaming"
)

func (s *Server) playbackSubtitle(w http.ResponseWriter, r *http.Request) {
	source, ok := s.playbackSource(w, r)
	if !ok {
		return
	}
	id, err := strconv.Atoi(r.PathValue("subtitle"))
	if err != nil {
		fail(w, 404, "not_found", "subtitle not found")
		return
	}
	var selected *streaming.Subtitle
	for _, subtitle := range streaming.Subtitles(source.Root, source.Item.RelativePath, source.Item.Tracks...) {
		if subtitle.ID == id {
			selected = &subtitle
			break
		}
	}
	if selected == nil {
		fail(w, 404, "not_found", "subtitle not found")
		return
	}
	if selected.Delivery == "burn" {
		fail(w, 422, "subtitle_requires_transcode", "bitmap subtitles require compatibility playback")
		return
	}
	position := 0.0
	if r.URL.Query().Has("start") {
		position, err = strconv.ParseFloat(r.URL.Query().Get("start"), 64)
		if err != nil || math.IsNaN(position) || math.IsInf(position, 0) || position < 0 || position > 86400 {
			fail(w, 400, "invalid_position", "invalid subtitle position")
			return
		}
	}
	var data []byte
	if selected.Kind == "embedded" {
		data, err = s.subtitles.Read(r.Context(), source.Path, -selected.ID-1, position)
	} else {
		data, err = streaming.ReadSubtitle(source.Root, selected.RelativePath)
	}
	if errors.Is(err, streaming.ErrBusy) {
		fail(w, 503, "subtitle_busy", "subtitle extraction busy; retry shortly")
		return
	}
	if err != nil {
		fail(w, 422, "subtitle_unavailable", "unable to read or convert this subtitle")
		return
	}
	w.Header().Set("Content-Type", "text/vtt; charset=utf-8")
	_, _ = w.Write(data)
}
