package server

import (
	"bufio"
	"context"
	"errors"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"dinosaur/internal/streaming"
)

// idleStreamWriter extends the connection deadline for each chunk, so a movie
// can outlive the JSON API timeout without retaining a stalled client forever.
type idleStreamWriter struct{ writer http.ResponseWriter }

func (w idleStreamWriter) Write(data []byte) (int, error) {
	_ = http.NewResponseController(w.writer).SetWriteDeadline(time.Now().Add(30 * time.Second))
	count, err := w.writer.Write(data)
	if err == nil {
		err = http.NewResponseController(w.writer).Flush()
	}
	return count, err
}
func (s *Server) servePlaybackStream(w http.ResponseWriter, r *http.Request) {
	source, ok := s.playbackSource(w, r)
	if !ok {
		return
	}
	plan, err := requestPlan(r, source.Item)
	if err != nil {
		fail(w, 400, "invalid_playback", err.Error())
		return
	}
	_ = http.NewResponseController(w).SetWriteDeadline(time.Time{})
	stop := context.AfterFunc(r.Context(), func() { _ = http.NewResponseController(w).SetWriteDeadline(time.Now()) })
	defer stop()
	if plan.Mode == "direct" {
		file, err := os.Open(source.Path)
		if err != nil {
			fail(w, 404, "media_unavailable", "media file unavailable")
			return
		}
		defer file.Close()
		info, err := file.Stat()
		if err != nil || !info.Mode().IsRegular() {
			fail(w, 404, "media_unavailable", "media file unavailable")
			return
		}
		kind := mime.TypeByExtension(filepath.Ext(source.Path))
		if kind == "" {
			kind = "application/octet-stream"
		}
		w.Header().Set("Content-Type", kind)
		http.ServeContent(streamResponse{ResponseWriter: w}, r, source.Item.Filename, info.ModTime(), file)
		return
	}
	w.Header().Set("Content-Type", "video/mp4")
	w.Header().Set("Accept-Ranges", "none")
	if r.Method == http.MethodHead {
		w.WriteHeader(200)
		return
	}
	stream, err := s.transcoder.Open(r.Context(), source.Path, plan)
	if err != nil {
		if errors.Is(err, streaming.ErrBusy) {
			fail(w, 503, "playback_busy", "two compatibility streams are already active")
			return
		}
		fail(w, 503, "playback_unavailable", "unable to start FFmpeg playback")
		s.logger.Warn("start playback failed", "media_id", source.Item.ID, "error", err)
		return
	}
	defer stream.Close()
	reader := bufio.NewReaderSize(stream, 64*1024)
	if _, err := reader.Peek(1); err != nil {
		if r.Context().Err() == nil {
			stream.Close()
			s.logger.Warn("FFmpeg playback initialization failed", "media_id", source.Item.ID, "details", stream.Diagnostics())
			fail(w, 422, "playback_failed", "FFmpeg could not read or convert this video")
		}
		return
	}
	stream.Ready()
	// fMP4 starts at time zero for each seek. The player tracks the source offset.
	_, err = io.CopyBuffer(idleStreamWriter{w}, reader, make([]byte, 64*1024))
	if err != nil && r.Context().Err() == nil {
		s.logger.Debug("playback stream closed", "media_id", source.Item.ID, "error", err)
	}
}

type streamResponse struct{ http.ResponseWriter }

func (w streamResponse) Write(data []byte) (int, error) {
	return idleStreamWriter{w.ResponseWriter}.Write(data)
}
func (w streamResponse) Unwrap() http.ResponseWriter { return w.ResponseWriter }
