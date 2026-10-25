package server

import (
	"net/http"
	"strconv"
	"strings"

	"dinosaur/internal/metadata"
)

func (s *Server) mountScrapeReports(mux *http.ServeMux) {
	mux.Handle("GET /api/admin/metadata/records", s.require(true, http.HandlerFunc(s.metadataRecords)))
	mux.Handle("GET /api/admin/metadata/jobs/{id}", s.require(true, http.HandlerFunc(s.scrapeJobDetail)))
}

func metadataReportFilter(w http.ResponseWriter, r *http.Request, pageSize int, states string) (metadata.ReportFilter, bool) {
	filter := metadata.ReportFilter{Page: 1, PageSize: pageSize, Status: r.URL.Query().Get("status"), Query: strings.TrimSpace(r.URL.Query().Get("q"))}
	validStatus := filter.Status == ""
	for _, allowed := range strings.Fields(states) {
		if filter.Status == allowed {
			validStatus = true
		}
	}
	if !validStatus {
		fail(w, 400, "invalid_filter", "invalid metadata status")
		return filter, false
	}
	if len(filter.Query) > 256 {
		fail(w, 400, "invalid_filter", "search query is too long")
		return filter, false
	}
	for key, pointer := range map[string]*int{"page": &filter.Page, "page_size": &filter.PageSize} {
		if value := r.URL.Query().Get(key); value != "" {
			number, err := strconv.Atoi(value)
			if err != nil || number < 1 || number > 1_000_000 || (key == "page_size" && number > 100) {
				fail(w, 400, "invalid_filter", "invalid pagination")
				return filter, false
			}
			*pointer = number
		}
	}
	if value := r.URL.Query().Get("library_id"); value != "" {
		id, err := strconv.ParseInt(value, 10, 64)
		if err != nil || id < 1 {
			fail(w, 400, "invalid_filter", "invalid library id")
			return filter, false
		}
		filter.LibraryID = id
	}
	return filter, true
}

func (s *Server) metadataRecords(w http.ResponseWriter, r *http.Request) {
	filter, ok := metadataReportFilter(w, r, 20, "matched unmatched needs_review failed warning")
	if !ok {
		return
	}
	page, err := s.store.MetadataRecords(r.Context(), filter)
	if s.scrapeError(w, r, err) {
		return
	}
	writeJSON(w, 200, page)
}

func (s *Server) scrapeJobDetail(w http.ResponseWriter, r *http.Request) {
	id, ok := libraryID(w, r)
	if !ok {
		return
	}
	filter, ok := metadataReportFilter(w, r, 20, "")
	if !ok {
		return
	}
	result, err := s.store.ScrapeJobDetail(r.Context(), id, filter)
	if s.scrapeError(w, r, err) {
		return
	}
	writeJSON(w, 200, result)
}
