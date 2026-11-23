package history

import (
	"errors"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"
)

var ErrInvalid = errors.New("invalid watch history record")
var ErrConflict = errors.New("watch history record already exists")
var digits = regexp.MustCompile(`^[1-9][0-9]{0,19}$`)
var key = regexp.MustCompile(`^[a-f0-9]{64}\.(webp|jpg|png)$`)

type Entry struct {
	ID            int64  `json:"id"`
	MediaID       *int64 `json:"media_id,omitempty"`
	Title         string `json:"title"`
	OriginalTitle string `json:"original_title"`
	Year          int    `json:"year"`
	WatchedOn     string `json:"watched_on"`
	Rating        int    `json:"rating"`
	Review        string `json:"review"`
	Source        string `json:"source"`
	SourceID      string `json:"source_id"`
	PosterKey     string `json:"-"`
	PosterURL     string `json:"poster_url,omitempty"`
	SourceURL     string `json:"source_url,omitempty"`
}

type Input struct {
	Title         string `json:"title"`
	OriginalTitle string `json:"original_title"`
	Year          int    `json:"year"`
	WatchedOn     string `json:"watched_on"`
	Rating        int    `json:"rating"`
	Review        string `json:"review"`
}

func (e *Entry) Apply(in Input) {
	e.Title, e.OriginalTitle = strings.TrimSpace(in.Title), strings.TrimSpace(in.OriginalTitle)
	e.Year, e.WatchedOn, e.Rating, e.Review = in.Year, in.WatchedOn, in.Rating, strings.TrimSpace(in.Review)
}

func (e Entry) Validate() error {
	day, err := time.Parse("2006-01-02", e.WatchedOn)
	if err != nil || day.Format("2006-01-02") != e.WatchedOn || day.Year() < 1900 {
		return ErrInvalid
	}
	if strings.TrimSpace(e.Title) == "" || utf8.RuneCountInString(e.Title) > 512 ||
		utf8.RuneCountInString(e.OriginalTitle) > 512 || utf8.RuneCountInString(e.Review) > 4000 ||
		e.Year < 0 || e.Year > 9999 || e.Rating < 0 || e.Rating > 5 || len(e.SourceID) == 0 || len(e.SourceID) > 128 {
		return ErrInvalid
	}
	if e.Source != "manual" && e.Source != "douban" && e.Source != "local" {
		return ErrInvalid
	}
	if e.Source == "douban" && !digits.MatchString(e.SourceID) {
		return ErrInvalid
	}
	if e.PosterKey != "" && !ValidPosterKey(e.PosterKey) {
		return ErrInvalid
	}
	return nil
}

func ValidPosterKey(value string) bool { return key.MatchString(value) }

func Day(t time.Time) string {
	return t.In(time.FixedZone("Asia/Shanghai", 8*3600)).Format("2006-01-02")
}

type Page struct {
	Items    []Entry `json:"items"`
	Total    int     `json:"total"`
	Page     int     `json:"page"`
	PageSize int     `json:"page_size"`
}

type ImportResult struct {
	Imported int `json:"imported"`
	Skipped  int `json:"skipped"`
}
