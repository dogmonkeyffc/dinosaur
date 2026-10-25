package metadata

// Job records durable progress for one background scraping operation.
type Job struct {
	MediaType string `json:"media_type"`
	ID        int64  `json:"id"`
	LibraryID int64  `json:"library_id"`
	MediaID   int64  `json:"media_id"`
	Status    string `json:"status"`
	Total     int    `json:"total"`
	Processed int    `json:"processed"`
	Matched   int    `json:"matched"`
	Review    int    `json:"review"`
	Errors    int    `json:"errors"`
	Skipped   int    `json:"skipped"`
	Message   string `json:"message"`
	CreatedAt int64  `json:"created_at"`
	EndedAt   int64  `json:"ended_at"`
}

type ScrapeRequest struct {
	MediaType string `json:"-"`
	LibraryID int64  `json:"library_id"`
	MediaID   int64  `json:"media_id"`
	TMDBID    int64  `json:"tmdb_id"`
	Refresh   bool   `json:"refresh"`
}
