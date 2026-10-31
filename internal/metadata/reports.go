package metadata

// Record is an administrator's view of current identification results.
type Record struct {
	SeriesID    int64       `json:"series_id,omitempty"`
	Candidates  []Candidate `json:"candidates,omitempty"`
	MediaID     int64       `json:"media_id"`
	LibraryID   int64       `json:"library_id"`
	LibraryName string      `json:"library_name"`
	Title       string      `json:"title"`
	Filename    string      `json:"filename"`
	BangumiID   int64       `json:"bangumi_id,omitempty"`
	TMDBID      int64       `json:"tmdb_id"`
	Language    string      `json:"language"`
	Status      string      `json:"status"`
	Message     string      `json:"message"`
	UpdatedAt   int64       `json:"updated_at"`
}

type ReportFilter struct {
	MediaType string
	LibraryID int64
	Query     string
	Status    string
	Page      int
	PageSize  int
}

type RecordPage struct {
	Items    []Record       `json:"items"`
	Total    int            `json:"total"`
	Summary  map[string]int `json:"summary"`
	Page     int            `json:"page"`
	PageSize int            `json:"page_size"`
}

// JobItem preserves the outcome at the time a job ran, even after rematching.
type JobItem struct {
	SeriesID  int64  `json:"series_id,omitempty"`
	ID        int64  `json:"id"`
	MediaID   int64  `json:"media_id"`
	Title     string `json:"title"`
	Filename  string `json:"filename"`
	Library   string `json:"library_name"`
	Status    string `json:"status"`
	Message   string `json:"message"`
	BangumiID int64  `json:"bangumi_id,omitempty"`
	TMDBID    int64  `json:"tmdb_id"`
	UpdatedAt int64  `json:"updated_at"`
}

type JobPage struct {
	Jobs     []Job `json:"jobs"`
	Total    int   `json:"total"`
	Page     int   `json:"page"`
	PageSize int   `json:"page_size"`
}

type JobDetail struct {
	Job      Job       `json:"job"`
	Items    []JobItem `json:"items"`
	Total    int       `json:"total"`
	Page     int       `json:"page"`
	PageSize int       `json:"page_size"`
}
