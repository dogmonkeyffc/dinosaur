package store

import (
	"context"
	"path/filepath"
	"strings"

	"dinosaur/internal/library"
)

func (s *Store) VideoLibraryRoot(ctx context.Context, id int64) (string, error) {
	var root string
	err := s.DB.QueryRowContext(ctx, `SELECT root_path FROM libraries WHERE id=? AND type='video' AND `+
		libraryVisible("libraries.id"), id, libraryViewer(ctx)).Scan(&root)
	return root, err
}

func (s *Store) VideoDirectoryIndex(ctx context.Context, id int64, directory string) (map[string]library.IndexedVideo, error) {
	prefix := filepath.FromSlash(directory)
	separator := string(filepath.Separator)
	if prefix != "" {
		prefix += separator
	}
	escaped := strings.NewReplacer(`\`, `\\`, "%", `\%`, "_", `\_`).Replace(prefix)
	rows, err := s.DB.QueryContext(ctx, `SELECT m.id,m.relative_path,m.poster_key FROM media_items m JOIN libraries l ON l.id=m.library_id
		WHERE m.library_id=? AND m.missing=0 AND l.type='video' AND m.relative_path LIKE ? ESCAPE '\'
		AND instr(substr(m.relative_path,length(?)+1),?)=0 AND `+libraryVisible("m.library_id"),
		id, escaped+"%", prefix, separator, libraryViewer(ctx))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	index := map[string]library.IndexedVideo{}
	for rows.Next() {
		var video library.IndexedVideo
		var relative string
		if err := rows.Scan(&video.ID, &relative, &video.PosterKey); err != nil {
			return nil, err
		}
		if name, ok := strings.CutPrefix(relative, prefix); ok {
			index[name] = video
		}
	}
	return index, rows.Err()
}
