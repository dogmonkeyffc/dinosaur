package server

import (
	"database/sql"
	"errors"
	"net/http"
	"strconv"

	"dinosaur/internal/television"
)

func (s *Server) mountTelevision(mux *http.ServeMux) {
	mux.Handle("GET /api/series", s.require(false, http.HandlerFunc(s.listSeries)))
	mux.Handle("GET /api/series/{id}", s.require(false, http.HandlerFunc(s.seriesDetail)))
	mux.Handle("GET /api/series/{id}/poster", s.require(false, http.HandlerFunc(s.seriesPoster)))
}

func setSeriesPoster(item *television.Series) {
	if validPosterKey(item.PosterKey) {
		item.PosterURL = "/api/series/" + strconv.FormatInt(item.ID, 10) + "/poster?v=" + item.PosterKey
	}
}

func (s *Server) listSeries(w http.ResponseWriter, r *http.Request) {
	filter, ok := mediaFilter(w, r)
	if !ok {
		return
	}
	page, err := s.store.SeriesPage(r.Context(), filter, current(r).user.ID)
	if err != nil {
		s.internal(w, r, err)
		return
	}
	for i := range page.Items {
		setSeriesPoster(&page.Items[i])
	}
	writeJSON(w, 200, page)
}

func (s *Server) seriesDetail(w http.ResponseWriter, r *http.Request) {
	id, ok := libraryID(w, r)
	if !ok {
		return
	}
	item, err := s.store.SeriesByID(r.Context(), id)
	if s.seriesError(w, r, err) {
		return
	}
	episodes, err := s.store.SeriesEpisodes(r.Context(), id, current(r).user.ID)
	if err != nil {
		s.internal(w, r, err)
		return
	}
	setSeriesPoster(&item)
	result := television.Detail{Series: item, Seasons: []television.Season{}}
	for i := range episodes {
		setPosterURL(&episodes[i])
		if episodes[i].Watched {
			result.Series.WatchedCount++
		}
		number := episodes[i].SeasonNumber
		if len(result.Seasons) == 0 || result.Seasons[len(result.Seasons)-1].Number != number {
			result.Seasons = append(result.Seasons, television.Season{Number: number})
		}
		season := &result.Seasons[len(result.Seasons)-1]
		season.Episodes = append(season.Episodes, episodes[i])
	}
	result.Resume = television.ResumeEpisode(episodes)
	writeJSON(w, 200, result)
}

func (s *Server) seriesPoster(w http.ResponseWriter, r *http.Request) {
	id, ok := libraryID(w, r)
	if !ok {
		return
	}
	item, err := s.store.SeriesByID(r.Context(), id)
	if s.seriesError(w, r, err) {
		return
	}
	s.servePoster(w, r, item.PosterKey)
}

func (s *Server) seriesError(w http.ResponseWriter, r *http.Request, err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, sql.ErrNoRows) {
		fail(w, 404, "not_found", "series not found")
	} else {
		s.internal(w, r, err)
	}
	return true
}
