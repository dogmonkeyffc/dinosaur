package history

type LibrarySetting struct {
	ID            int64  `json:"id"`
	Name          string `json:"name"`
	Type          string `json:"type"`
	RecordHistory bool   `json:"record_history"`
}
