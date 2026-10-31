package server

import (
	"net/http"

	"dinosaur/internal/metadata"
)

func (s *Server) mountTelevisionScraping(mux *http.ServeMux) {
	mux.Handle("POST /api/admin/tv/metadata/jobs", s.require(true, http.HandlerFunc(s.startSeriesScrape)))
	mux.Handle("GET /api/admin/tv/metadata/jobs", s.require(true, http.HandlerFunc(s.seriesScrapeJobs)))
	mux.Handle("GET /api/admin/series/{id}/metadata/candidates", s.require(true, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.scrapeCandidatesWith(w, r, s.tvScraper)
	})))
}

func (s *Server) startSeriesScrape(w http.ResponseWriter, r *http.Request) {
	var body struct {
		LibraryID int64 `json:"library_id"`
		SeriesID  int64 `json:"series_id"`
		BangumiID int64 `json:"bangumi_id"`
		TMDBID    int64 `json:"tmdb_id"`
		Refresh   bool  `json:"refresh"`
	}
	if !decodeJSON(w, r, &body) {
		return
	}
	request := metadata.ScrapeRequest{LibraryID: body.LibraryID, MediaID: body.SeriesID, TMDBID: body.TMDBID, BangumiID: body.BangumiID, Refresh: body.Refresh}
	job, err := s.tvScraper.Start(r.Context(), request)
	if s.scrapeError(w, r, err) {
		return
	}
	writeJSON(w, 202, map[string]any{"job": job})
}

func (s *Server) seriesScrapeJobs(w http.ResponseWriter, r *http.Request) {
	filter, ok := metadataReportFilter(w, r, 10, "queued running completed failed interrupted")
	if !ok {
		return
	}
	filter.MediaType = "tv"
	page, err := s.store.ScrapeJobPage(r.Context(), filter)
	if s.scrapeError(w, r, err) {
		return
	}
	writeJSON(w, 200, page)
}
