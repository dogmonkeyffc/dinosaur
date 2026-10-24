package library

// Library configures a root directory containing one kind of media.
type Library struct {
	ID         int64  `json:"id"`
	Name       string `json:"name"`
	Type       string `json:"type"`
	RootPath   string `json:"root_path"`
	CreatedAt  int64  `json:"created_at"`
	UpdatedAt  int64  `json:"updated_at"`
	MediaCount int    `json:"media_count"`
}

// ScanJob is the durable progress of one complete library traversal.
type ScanJob struct {
	ID         int64  `json:"id"`
	LibraryID  int64  `json:"library_id"`
	Status     string `json:"status"`
	Discovered int    `json:"discovered"`
	Processed  int    `json:"processed"`
	Errors     int    `json:"errors"`
	Message    string `json:"message"`
	CreatedAt  int64  `json:"created_at"`
	FinishedAt int64  `json:"finished_at"`
}
