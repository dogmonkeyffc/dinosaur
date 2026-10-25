package server

import (
	"errors"
	"net/http"

	"dinosaur/internal/disk"
)

func (s *Server) mountSystem(mux *http.ServeMux) {
	mux.Handle("GET /api/admin/system/disks", s.require(true, http.HandlerFunc(s.diskUsage)))
}

func (s *Server) diskUsage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	report, err := s.disks.Read(r.Context())
	if errors.Is(err, disk.ErrUnsupported) {
		fail(w, 501, "disk_usage_unsupported", "disk usage is not supported on this platform")
		return
	}
	if err != nil {
		s.internal(w, r, err)
		return
	}
	writeJSON(w, 200, report)
}
