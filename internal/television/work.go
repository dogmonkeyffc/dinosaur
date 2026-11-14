package television

import (
	"strings"
	"unicode/utf8"
)

// Work groups series without modifying their scan identity or metadata.
type Work struct {
	ID           int64  `json:"id"`
	Title        string `json:"title"`
	LibraryID    int64  `json:"library_id"`
	PosterURL    string `json:"poster_url,omitempty"`
	EpisodeCount int    `json:"episode_count"`
	SeasonCount  int    `json:"season_count"`
	WatchedCount int    `json:"watched_count"`
	MemberCount  int    `json:"member_count"`
}

type WorkDetail struct {
	Work       Work     `json:"work"`
	Members    []Series `json:"members"`
	AllMembers []Series `json:"all_members,omitempty"`
}

type WorkError struct{ Code, Message string }

func (e *WorkError) Error() string { return e.Message }

func ValidateWork(title string, ids []int64) (string, error) {
	title = strings.TrimSpace(title)
	if title == "" || len(ids) == 0 {
		return "", &WorkError{"invalid_work", "title and at least one member are required"}
	}
	if utf8.RuneCountInString(title) > 128 || len(ids) > 256 {
		return "", &WorkError{"invalid_work", "title must not exceed 128 characters and membership must not exceed 256 series"}
	}
	seen := map[int64]bool{}
	for _, id := range ids {
		if id <= 0 || seen[id] {
			return "", &WorkError{"invalid_work", "member IDs must be positive and unique"}
		}
		seen[id] = true
	}
	return title, nil
}

// ValidateWorkMember enforces library and exclusive-membership invariants.
func ValidateWorkMember(libraryID, memberLibrary, workID, memberWork int64) error {
	if libraryID != 0 && libraryID != memberLibrary {
		return &WorkError{"cross_library", "all members must belong to the same library"}
	}
	if memberWork != 0 && memberWork != workID {
		return &WorkError{"work_conflict", "series already belongs to another work"}
	}
	return nil
}
