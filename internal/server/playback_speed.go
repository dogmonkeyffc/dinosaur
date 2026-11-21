package server

import (
	"context"
	"net/http"
	"time"

	"dinosaur/internal/playmetrics"
	"dinosaur/internal/streaming"
)

// Track only browser-correlated transcodes. Stop sampling before the final
// write so an older sample cannot overwrite the final speed.
func (s *Server) trackTranscodeSpeed(r *http.Request, stream *streaming.Stream) func() {
	id := r.URL.Query().Get("stream_id")
	who := current(r)
	if who.device || !playmetrics.ValidID(id) {
		return func() {}
	}
	ctx, stop := context.WithCancel(r.Context())
	done := make(chan struct{})
	var saved float64
	record := func(ctx context.Context) {
		speed := stream.TranscodeSpeed()
		if speed <= 0 || speed == saved {
			return
		}
		ctx, cancel := context.WithTimeout(ctx, time.Second)
		defer cancel()
		if s.store.RecordTranscodeSpeed(ctx, who.user.ID, id, speed) == nil {
			saved = speed
		}
	}
	go func() {
		defer close(done)
		ticker := time.NewTicker(15 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				record(ctx)
			}
		}
	}()
	return func() {
		stop()
		<-done
		record(context.WithoutCancel(r.Context()))
	}
}
