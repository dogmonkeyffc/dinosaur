package metadata

import (
	"strings"
	"unicode"
)

// Movie is provider information, independent of media files and HTTP DTOs.
type Movie struct {
	Candidates    []Candidate `json:"candidates,omitempty"`
	TMDBID        int64       `json:"tmdb_id"`
	Title         string      `json:"title"`
	OriginalTitle string      `json:"original_title"`
	Overview      string      `json:"overview"`
	Year          int         `json:"year"`
	ReleaseDate   string      `json:"release_date"`
	Language      string      `json:"language"`
	Status        string      `json:"status"`
	Message       string      `json:"message"`
	UpdatedAt     int64       `json:"updated_at"`
	PosterKey     string      `json:"-"`
	PosterPath    string      `json:"-"`
	Aliases       string      `json:"-"`
}

type Candidate struct {
	ID            int64  `json:"id"`
	Title         string `json:"title"`
	OriginalTitle string `json:"original_title"`
	Year          int    `json:"year"`
	Overview      string `json:"overview"`
}

func normalizedTitle(value string) string {
	return strings.Map(func(char rune) rune {
		if unicode.IsLetter(char) || unicode.IsDigit(char) {
			return unicode.ToLower(char)
		}
		return -1
	}, value)
}

// Automatic matching requires a known year and one unambiguous candidate.
// A one-year release discrepancy also requires an exact normalized title.
func AutoMatch(query string, year int, candidates []Candidate) (int64, bool) {
	if year == 0 || normalizedTitle(query) == "" {
		return 0, false
	}
	matches := make(map[int64]bool)
	for _, item := range candidates {
		exact := normalizedTitle(query) == normalizedTitle(item.Title) || normalizedTitle(query) == normalizedTitle(item.OriginalTitle)
		if item.ID > 0 && (item.Year == year || (exact && (item.Year == year-1 || item.Year == year+1))) {
			matches[item.ID] = true
		}
	}
	if len(matches) == 1 {
		for id := range matches {
			return id, true
		}
	}
	return 0, false
}
