// Package agent defines scoped metadata integrations independently of HTTP and SQL.
package agent

import (
	"dinosaur/internal/media"
	"dinosaur/internal/metadata"
	"errors"
)

var ErrInput = errors.New("invalid agent metadata input")
var ErrMetadataRequired = errors.New("confirm metadata before uploading artwork")
var ErrForbidden = errors.New("agent key does not permit this operation")

type Key struct {
	ID        int64  `json:"id"`
	UserID    int64  `json:"-"`
	LibraryID int64  `json:"library_id"`
	Name      string `json:"name"`
	Prefix    string `json:"prefix"`
	Access    string `json:"access"`
	CreatedAt int64  `json:"created_at"`
	ExpiresAt int64  `json:"expires_at"`
}

type CreateKey struct {
	Name          string `json:"name"`
	LibraryID     int64  `json:"library_id"`
	Access        string `json:"access"`
	ExpiresInDays int    `json:"expires_in_days"`
}

func (k Key) Allows(libraryID int64, write bool) bool {
	return (k.LibraryID == 0 || k.LibraryID == libraryID) && (!write || k.Access == "read_write")
}

type File struct {
	Path       string        `json:"path"`
	Size       int64         `json:"size"`
	Duration   float64       `json:"duration"`
	VideoCodec string        `json:"video_codec"`
	Width      int           `json:"width"`
	Height     int           `json:"height"`
	Tracks     []media.Track `json:"tracks"`
	Season     int           `json:"season,omitempty"`
	Episode    int           `json:"episode,omitempty"`
}

type Target struct {
	Kind        string          `json:"kind"`
	ID          int64           `json:"id"`
	LibraryID   int64           `json:"library_id"`
	LibraryName string          `json:"library_name"`
	LibraryRoot string          `json:"library_root"`
	Title       string          `json:"title"`
	Year        int             `json:"year"`
	Version     int64           `json:"version"`
	FileCount   int             `json:"file_count"`
	Files       []File          `json:"files"`
	Metadata    *metadata.Movie `json:"metadata,omitempty"`
}
