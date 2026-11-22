package server

import (
	"errors"
	"net/http"
	"path/filepath"

	"dinosaur/internal/library"
	"dinosaur/internal/media"
)

func (s *Server) browseLibraryDirectories(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	directory := query.Get("path")
	if directory == "" {
		directory = string(filepath.Separator)
		if library.Readable("/mnt") {
			directory = "/mnt"
		}
	}
	if !filepath.IsAbs(directory) || len(directory) > 4096 || filepath.Clean(directory) != directory {
		fail(w, 400, "invalid_library_path", "specify an absolute directory path")
		return
	}
	root := filepath.VolumeName(directory) + string(filepath.Separator)
	relative, err := filepath.Rel(root, directory)
	if err != nil {
		fail(w, 400, "invalid_library_path", "specify an absolute directory path")
		return
	}
	entries, next, err := media.FolderPage(r.Context(), root, filepath.ToSlash(relative), query.Get("after"))
	if errors.Is(err, media.ErrInvalidDirectory) {
		fail(w, 400, "invalid_library_path", "invalid directory")
		return
	}
	if err != nil {
		fail(w, 404, "library_path_unavailable", "directory cannot be browsed")
		return
	}
	type folder struct {
		Name string `json:"name"`
		Path string `json:"path"`
	}
	folders := make([]folder, 0, len(entries))
	for _, entry := range entries {
		folders = append(folders, folder{filepath.Base(entry.Path), filepath.Join(root, entry.Path)})
	}
	parent := filepath.Dir(directory)
	if parent == directory {
		parent = ""
	}
	w.Header().Set("Cache-Control", "no-store")
	writeJSON(w, 200, map[string]any{"path": directory, "parent": parent, "folders": folders, "next_after": next})
}
