package server

import (
	"database/sql"
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"dinosaur/internal/media"
)

func (s *Server) listMedia(w http.ResponseWriter, r *http.Request) {
	filter, ok := mediaFilter(w, r)
	if !ok {
		return
	}
	filter.ViewerID = current(r).user.ID
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
	s.servePoster(w, r, item.PosterKey)
}
func (s *Server) servePoster(w http.ResponseWriter, r *http.Request, key string) {
	if !validPosterKey(key) {
		fail(w, 404, "not_found", "poster unavailable")
		return
	}
	path := filepath.Join(s.config.DataDir, "posters", key+".jpg")
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
	w.Header().Set("Cache-Control", "private, no-store")
	w.Header().Set("Content-Type", "image/jpeg")
	http.ServeContent(w, r, "poster.jpg", info.ModTime(), file)
}
func setPosterURL(item *media.Item) {
	item.PosterURL = mediaPosterURL(item.ID, item.PosterKey)
}

func mediaPosterURL(id int64, key string) string {
	if !validPosterKey(key) {
		return ""
	}
	return "/api/media/" + strconv.FormatInt(id, 10) + "/poster?v=" + key
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
