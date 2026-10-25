// Package television defines local series and season browsing independently of storage.
package television

import (
	"dinosaur/internal/media"
	"dinosaur/internal/metadata"
)

type EpisodeName struct {
	Key     string
	Series  string
	Year    int
	Season  int
	Episode int
	End     int
	Title   string
}

type Series struct {
	ID           int64           `json:"id"`
	LibraryID    int64           `json:"library_id"`
	LibraryName  string          `json:"library_name"`
	Title        string          `json:"title"`
	Year         int             `json:"year"`
	EpisodeCount int             `json:"episode_count"`
	SeasonCount  int             `json:"season_count"`
	WatchedCount int             `json:"watched_count"`
	PosterKey    string          `json:"-"`
	PosterURL    string          `json:"poster_url,omitempty"`
	Metadata     *metadata.Movie `json:"metadata,omitempty"`
}

type Season struct {
	Number   int          `json:"number"`
	Episodes []media.Item `json:"episodes"`
}

type Detail struct {
	Series  Series      `json:"series"`
	Seasons []Season    `json:"seasons"`
	Resume  *media.Item `json:"resume,omitempty"`
}

type Page struct {
	Items    []Series `json:"items"`
	Total    int      `json:"total"`
	Page     int      `json:"page"`
	PageSize int      `json:"page_size"`
}
