package server

import (
	"database/sql"
	"errors"
	"io/fs"
	"net/http"
	"path"
	"path/filepath"
	"strconv"

	"dinosaur/internal/library"
)

func (s *Server) videoDirectory(w http.ResponseWriter, r *http.Request) {
	id, ok := libraryID(w, r)
	if !ok {
		return
	}
	ctx := r.Context()
	root, err := s.store.VideoLibraryRoot(ctx, id)
	if s.videoDirectoryError(w, r, err) {
		return
	}
	directory := r.URL.Query().Get("path")
	if rawID := r.URL.Query().Get("media_id"); rawID != "" {
		mediaID, err := strconv.ParseInt(rawID, 10, 64)
		if err != nil || mediaID < 1 || directory != "" {
			s.videoDirectoryError(w, r, fs.ErrInvalid)
			return
		}
		item, err := s.store.MediaByID(ctx, mediaID)
		if s.videoDirectoryError(w, r, err) {
			return
		}
		if item.LibraryID != id {
			s.videoDirectoryError(w, r, fs.ErrNotExist)
			return
		}
		directory = path.Dir(filepath.ToSlash(item.RelativePath))
		if directory == "." {
			directory = ""
		}
	}
	if directory != "" && (directory == "." || !fs.ValidPath(directory)) {
		s.videoDirectoryError(w, r, fs.ErrInvalid)
		return
	}
	index, err := s.store.VideoDirectoryIndex(ctx, id, directory)
	if s.videoDirectoryError(w, r, err) {
		return
	}
	result, err := library.ReadVideoDirectory(root, directory, index)
	if !s.videoDirectoryError(w, r, err) {
		writeJSON(w, http.StatusOK, result)
	}
}

func (s *Server) videoDirectoryError(w http.ResponseWriter, r *http.Request, err error) bool {
	switch {
	case err == nil:
		return false
	case errors.Is(err, fs.ErrInvalid):
		fail(w, http.StatusBadRequest, "invalid_directory", "invalid directory")
	case errors.Is(err, sql.ErrNoRows), errors.Is(err, fs.ErrNotExist), errors.Is(err, fs.ErrPermission):
		fail(w, http.StatusNotFound, "not_found", "directory unavailable")
	default:
		s.internal(w, r, err)
	}
	return true
}
