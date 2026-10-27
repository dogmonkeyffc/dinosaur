package television

import (
	"path/filepath"
	"sort"

	"dinosaur/internal/media"
)

// EpisodeFolders lists distinct library-relative directories of present episodes.
func EpisodeFolders(episodes []media.Item) []string {
	folders := make([]string, 0)
	seen := make(map[string]bool)
	for _, episode := range episodes {
		if episode.Missing || episode.RelativePath == "" {
			continue
		}
		folder := filepath.ToSlash(filepath.Dir(episode.RelativePath))
		if !seen[folder] {
			seen[folder] = true
			folders = append(folders, folder)
		}
	}
	sort.Strings(folders)
	return folders
}
