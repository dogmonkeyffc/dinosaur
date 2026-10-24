package server

import (
	"database/sql"
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"dinosaur/internal/media"
)

func (s *Server) listMedia(w http.ResponseWriter, r *http.Request) {
	filter := media.Filter{Page: 1, PageSize: 24, Query: strings.TrimSpace(r.URL.Query().Get("q"))}
	for name, target := range map[string]*int{"page": &filter.Page, "page_size": &filter.PageSize} {
		value := r.URL.Query().Get(name)
		if value == "" {
			continue
		}
		number, err := strconv.Atoi(value)
		if err != nil || number < 1 || number > 1_000_000 || (name == "page_size" && number > 100) {
			fail(w, 400, "invalid_filter", "invalid pagination parameters")
			return
		}
		*target = number
	}
	if value := r.URL.Query().Get("library_id"); value != "" {
		id, err := strconv.ParseInt(value, 10, 64)
		if err != nil || id <= 0 {
			fail(w, 400, "invalid_filter", "invalid library ID")
			return
		}
		filter.LibraryID = id
	}
	if len(filter.Query) > 256 {
		fail(w, 400, "invalid_filter", "search query is too long")
		return
	}
	page, err := s.store.MediaPage(r.Context(), filter)
	if err != nil {
		s.internal(w, r, err)
		return
	}
	if err := s.store.AddPlaybackProgress(r.Context(), current(r).user.ID, page.Items); err != nil {
		s.internal(w, r, err)
		return
	}
	for i := range page.Items {
		setPosterURL(&page.Items[i])
	}
	writeJSON(w, 200, page)
}
func (s *Server) getMedia(w http.ResponseWriter, r *http.Request) {
	id, ok := libraryID(w, r)
	if !ok {
		return
	}
	item, err := s.store.MediaByID(r.Context(), id)
	if errors.Is(err, sql.ErrNoRows) {
		fail(w, 404, "not_found", "media not found")
		return
	}
	if err != nil {
		s.internal(w, r, err)
		return
	}
	items := []media.Item{item}
	if err := s.store.AddPlaybackProgress(r.Context(), current(r).user.ID, items); err != nil {
		s.internal(w, r, err)
		return
	}
	item = items[0]
	setPosterURL(&item)
	writeJSON(w, 200, map[string]any{"item": item})
}
func (s *Server) mediaPoster(w http.ResponseWriter, r *http.Request) {
	id, ok := libraryID(w, r)
	if !ok {
		return
	}
	item, err := s.store.MediaByID(r.Context(), id)
	if errors.Is(err, sql.ErrNoRows) {
		fail(w, 404, "not_found", "media not found")
		return
	}
	if err != nil {
		s.internal(w, r, err)
		return
	}
	if !validPosterKey(item.PosterKey) {
		fail(w, 404, "not_found", "poster unavailable")
		return
	}
	path := filepath.Join(s.config.DataDir, "posters", item.PosterKey+".jpg")
	file, err := os.Open(path)
	if err != nil {
		fail(w, 404, "not_found", "poster unavailable")
		return
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil || !info.Mode().IsRegular() {
		fail(w, 404, "not_found", "poster unavailable")
		return
	}
	w.Header().Set("Content-Type", "image/jpeg")
	http.ServeContent(w, r, "poster.jpg", info.ModTime(), file)
}
func setPosterURL(item *media.Item) {
	if validPosterKey(item.PosterKey) {
		item.PosterURL = "/api/media/" + strconv.FormatInt(item.ID, 10) + "/poster?v=" + item.PosterKey
	}
}
func validPosterKey(key string) bool {
	if len(key) != 64 {
		return false
	}
	for _, c := range key {
		if !((c >= 'a' && c <= 'f') || (c >= '0' && c <= '9')) {
			return false
		}
	}
	return true
}
