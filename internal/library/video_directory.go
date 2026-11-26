package library

import (
	"io/fs"
	"os"
	"path/filepath"
)

type VideoFile struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type VideoDirectory struct {
	Path    string      `json:"path"`
	Folders []string    `json:"folders"`
	Videos  []VideoFile `json:"videos"`
}

// ReadVideoDirectory lists only real children of the selected directory and
// intersects its regular files with the media index. The API uses slash paths.
func ReadVideoDirectory(rootPath, directory string, indexed map[string]int64) (VideoDirectory, error) {
	result := VideoDirectory{Path: directory, Folders: []string{}, Videos: []VideoFile{}}
	name := directory
	if name == "" {
		name = "."
	}
	if !fs.ValidPath(name) {
		return result, fs.ErrInvalid
	}
	resolved, err := filepath.EvalSymlinks(rootPath)
	if err != nil {
		return result, err
	}
	if resolved != rootPath {
		return result, fs.ErrNotExist
	}
	root, err := os.OpenRoot(rootPath)
	if err != nil {
		return result, err
	}
	defer root.Close()
	entries, err := fs.ReadDir(root.FS(), name)
	if err != nil {
		return result, err
	}
	for _, entry := range entries {
		if entry.IsDir() {
			result.Folders = append(result.Folders, entry.Name())
		} else if id := indexed[entry.Name()]; id > 0 && entry.Type().IsRegular() {
			result.Videos = append(result.Videos, VideoFile{ID: id, Name: entry.Name()})
		}
	}
	return result, nil
}
