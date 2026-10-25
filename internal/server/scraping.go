package server

import (
	"database/sql"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"dinosaur/internal/auth"
	"dinosaur/internal/metadata"
	"dinosaur/internal/scraping"
)

func (s *Server) mountScraping(mux *http.ServeMux) {
	s.mountScrapeReports(mux)
	mux.Handle("POST /api/admin/metadata/jobs", s.require(true, http.HandlerFunc(s.startScrape)))
	mux.Handle("GET /api/admin/metadata/jobs", s.require(true, http.HandlerFunc(s.scrapeJobs)))
	mux.Handle("GET /api/admin/media/{id}/metadata/candidates", s.require(true, http.HandlerFunc(s.scrapeCandidates)))
}

func (s *Server) startScrape(w http.ResponseWriter, r *http.Request) {
	var body metadata.ScrapeRequest
	if !decodeJSON(w, r, &body) {
		return
	}
	job, err := s.scraper.Start(r.Context(), body)
	if s.scrapeError(w, r, err) {
		return
	}
	writeJSON(w, 202, map[string]any{"job": job})
}

func (s *Server) scrapeJobs(w http.ResponseWriter, r *http.Request) {
	filter, ok := metadataReportFilter(w, r, 10, "queued running completed failed interrupted")
	if !ok {
		return
	}
	page, err := s.store.ScrapeJobPage(r.Context(), filter)
	if s.scrapeError(w, r, err) {
		return
	}
	writeJSON(w, 200, page)
}

func (s *Server) scrapeCandidates(w http.ResponseWriter, r *http.Request) {
	id, ok := libraryID(w, r)
	if !ok {
		return
	}
	if !s.rate(w, auth.Limit{Key: "tmdb-search:" + strconv.FormatInt(current(r).user.ID, 10), Max: 60}) {
		return
	}
	query := strings.TrimSpace(r.URL.Query().Get("q"))
	if len(query) > 256 || (r.URL.Query().Has("q") && query == "") {
		fail(w, 400, "invalid_metadata_query", "invalid movie search query")
		return
	}
	var year *int
	if value := r.URL.Query().Get("year"); value != "" {
		number, err := strconv.Atoi(value)
		if err != nil || number < 0 || number > 2100 {
			fail(w, 400, "invalid_metadata_query", "invalid release year")
			return
		}
		year = &number
	}
	result, err := s.scraper.Candidates(r.Context(), id, query, year)
	if s.scrapeError(w, r, err) {
		return
	}
	writeJSON(w, 200, result)
}

func (s *Server) scrapeError(w http.ResponseWriter, r *http.Request, err error) bool {
	switch {
	case err == nil:
		return false
	case errors.Is(err, sql.ErrNoRows), errors.Is(err, metadata.ErrProviderNotFound):
		fail(w, 404, "not_found", "movie or library not found")
	case errors.Is(err, metadata.ErrMissingCredential):
		fail(w, 400, "tmdb_not_configured", "TMDB credential is not configured")
	case errors.Is(err, metadata.ErrProviderAuth):
		fail(w, 422, "tmdb_invalid_credential", "TMDB rejected the credential")
	case errors.Is(err, metadata.ErrProviderRateLimited):
		fail(w, 503, "tmdb_rate_limited", "TMDB rate limit exceeded")
	case errors.Is(err, metadata.ErrProviderUnavailable):
		fail(w, 502, "tmdb_unavailable", "cannot connect to TMDB")
	case errors.Is(err, scraping.ErrRunning):
		fail(w, 409, "scrape_running", "a scraping job is already running")
	case errors.Is(err, scraping.ErrRequest):
		fail(w, 400, "invalid_scrape", "invalid scraping request")
	case errors.Is(err, scraping.ErrClosed):
		fail(w, 503, "scraper_unavailable", "scraper is shutting down")
	default:
		s.internal(w, r, err)
	}
	return true
}
