package library

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"
)

type ValidationError struct{ Code, Message string }

func (e *ValidationError) Error() string { return e.Message }

// Validate resolves symlinks so aliases cannot register the same root twice.
// This only reads the root; it never creates or modifies the media directory.
func Validate(item Library) (Library, error) {
	item.Name = strings.TrimSpace(item.Name)
	if !utf8.ValidString(item.Name) || utf8.RuneCountInString(item.Name) < 1 || utf8.RuneCountInString(item.Name) > 80 {
		return item, &ValidationError{"invalid_library_name", "library name must contain 1-80 characters"}
	}
	if item.Type != "movie" && item.Type != "tv" {
		return item, &ValidationError{"invalid_library_type", "type must be movie or tv"}
	}
	if !filepath.IsAbs(item.RootPath) || len(item.RootPath) > 4096 {
		return item, &ValidationError{"invalid_library_path", "media directory must be an absolute server path"}
	}
	root, err := filepath.EvalSymlinks(filepath.Clean(item.RootPath))
	if err != nil || !Readable(root) {
		return item, &ValidationError{"library_path_unavailable", "media directory must exist and be readable by the server"}
	}
	item.RootPath = root
	return item, nil
}
func Readable(root string) bool {
	file, err := os.Open(root)
	if err != nil {
		return false
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil || !info.IsDir() {
		return false
	}
	_, err = file.Readdirnames(1)
	return err == nil || errors.Is(err, io.EOF)
}
