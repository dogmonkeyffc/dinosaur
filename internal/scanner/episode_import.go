package scanner

import (
	"path/filepath"

	"dinosaur/internal/media"
	"dinosaur/internal/television"
)

func validateAssignments(root string, assignments []television.EpisodeAssignment) error {
	if len(assignments) < 1 || len(assignments) > 100 {
		return television.ErrEpisodeAssignment
	}
	seen := make(map[string]bool, len(assignments))
	for _, assignment := range assignments {
		if _, err := assignment.Name(); err != nil || seen[assignment.Path] {
			return television.ErrEpisodeAssignment
		}
		seen[assignment.Path] = true
		resolved, err := media.ResolveFile(root, assignment.Path)
		if err != nil || resolved != filepath.Join(root, assignment.Path) {
			return television.ErrEpisodeAssignment
		}
	}
	return nil
}
