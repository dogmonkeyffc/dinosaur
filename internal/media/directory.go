package media

import (
	"context"
	"errors"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"
)

var ErrInvalidDirectory = errors.New("invalid media directory path")

type DirectoryEntry struct {
	Path string `json:"path"`
	Kind string `json:"kind"`
}

const directoryPageSize = 100

// DirectoryPage lists one level of the live filesystem within the configured root.
func DirectoryPage(ctx context.Context, root, directory, after string) ([]DirectoryEntry, string, error) {
	if directory == "" {
		directory = "."
	}
	if !fs.ValidPath(directory) || strings.Contains(directory, "\\") || strings.ContainsAny(after, "/\\") {
		return nil, "", ErrInvalidDirectory
	}
	canonical, err := filepath.EvalSymlinks(root)
	if err != nil {
		return nil, "", err
	}
	if canonical != filepath.Clean(root) {
		return nil, "", fs.ErrNotExist
	}
	original, err := os.Lstat(root)
	if err != nil {
		return nil, "", err
	}
	dir, err := os.OpenRoot(root)
	if err != nil {
		return nil, "", err
	}
	defer dir.Close()
	opened, err := dir.Stat(".")
	if err != nil {
		return nil, "", err
	}
	if !os.SameFile(original, opened) {
		return nil, "", fs.ErrNotExist
	}
	if directory != "." {
		current := "."
		for _, name := range strings.Split(directory, "/") {
			if strings.HasPrefix(name, ".") {
				return nil, "", ErrInvalidDirectory
			}
			current = path.Join(current, name)
			info, err := dir.Lstat(current)
			if err != nil {
				return nil, "", err
			}
			if !info.IsDir() {
				return nil, "", fs.ErrNotExist
			}
		}
	}
	entries, err := fs.ReadDir(dir.FS(), directory)
	if err != nil {
		return nil, "", err
	}
	items := make([]DirectoryEntry, 0, min(directoryPageSize, len(entries)))
	next := ""
	for _, entry := range entries {
		if err := ctx.Err(); err != nil {
			return nil, "", err
		}
		name := entry.Name()
		if name <= after || strings.HasPrefix(name, ".") || strings.Contains(name, "\\") {
			continue
		}
		kind := "directory"
		if !entry.IsDir() {
			if !entry.Type().IsRegular() || !IsVideo(name) {
				continue
			}
			kind = "video"
		}
		if len(items) == directoryPageSize {
			next = path.Base(items[len(items)-1].Path)
			break
		}
		items = append(items, DirectoryEntry{Path: path.Join(directory, name), Kind: kind})
	}
	return items, next, nil
}
