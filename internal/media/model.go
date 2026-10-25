package media

import "dinosaur/internal/metadata"

// Track contains the playable stream metadata, without arbitrary container tags.
type Track struct {
	Index    int    `json:"index"`
	Type     string `json:"type"`
	Codec    string `json:"codec"`
	Language string `json:"language"`
	Title    string `json:"title"`
	Channels int    `json:"channels,omitempty"`
}

type Item struct {
	Metadata     *metadata.Movie `json:"metadata,omitempty"`
	Position     float64         `json:"position"`
	Watched      bool            `json:"watched"`
	LastPlayedAt int64           `json:"last_played_at"`
	ID           int64           `json:"id"`
	LibraryID    int64           `json:"library_id"`
	LibraryName  string          `json:"library_name"`
	Title        string          `json:"title"`
	Year         int             `json:"year"`
	Filename     string          `json:"filename"`
	Size         int64           `json:"size"`
	ModifiedNS   int64           `json:"-"`
	RelativePath string          `json:"-"`
	Duration     float64         `json:"duration"`
	Format       string          `json:"format"`
	VideoCodec   string          `json:"video_codec"`
	Width        int             `json:"width"`
	Height       int             `json:"height"`
	Tracks       []Track         `json:"tracks"`
	PosterKey    string          `json:"-"`
	PosterURL    string          `json:"poster_url,omitempty"`
	ProbeError   bool            `json:"probe_error"`
	Missing      bool            `json:"missing"`
}

type Filter struct {
	LibraryID int64
	Query     string
	Page      int
	PageSize  int
}

type Page struct {
	Items    []Item `json:"items"`
	Total    int    `json:"total"`
	Page     int    `json:"page"`
	PageSize int    `json:"page_size"`
}
