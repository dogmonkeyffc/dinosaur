package agent

import (
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"

	"dinosaur/internal/metadata"
)

var languagePattern = regexp.MustCompile(`^[a-z]{2,3}(-[A-Za-z0-9]{2,8})*$`)
var idPattern = regexp.MustCompile(`^[a-z0-9_]{1,24}$`)

type Input struct {
	Title            string            `json:"title"`
	OriginalTitle    string            `json:"original_title"`
	Overview         string            `json:"overview"`
	Year             int               `json:"year"`
	ReleaseDate      string            `json:"release_date"`
	Language         string            `json:"language"`
	OriginalLanguage string            `json:"original_language"`
	Aliases          []string          `json:"aliases"`
	IDs              map[string]string `json:"ids"`
	SourceURL        string            `json:"source_url"`
}

func validText(value string, max int, multiline bool) bool {
	return len(value) <= max && !strings.ContainsFunc(value, func(r rune) bool {
		return unicode.IsControl(r) && !(multiline && (r == '\n' || r == '\r' || r == '\t'))
	})
}

func (input Input) Movie(source string) (metadata.Movie, error) {
	input.Title = strings.TrimSpace(input.Title)
	if input.Title == "" || !validText(input.Title, 512, false) || !validText(input.OriginalTitle, 512, false) || !validText(input.Overview, 32000, true) || input.Year < 0 || input.Year > 2200 || len(input.Aliases) > 50 || len(input.IDs) > 20 {
		return metadata.Movie{}, ErrInput
	}
	for _, language := range []string{input.Language, input.OriginalLanguage} {
		if len(language) > 32 || (language != "" && !languagePattern.MatchString(language)) {
			return metadata.Movie{}, ErrInput
		}
	}
	if input.ReleaseDate != "" {
		if _, err := time.Parse("2006-01-02", input.ReleaseDate); err != nil {
			return metadata.Movie{}, ErrInput
		}
	}
	if input.SourceURL != "" {
		u, err := url.Parse(input.SourceURL)
		if err != nil || len(input.SourceURL) > 2048 || (u.Scheme != "https" && u.Scheme != "http") || u.Hostname() == "" || u.User != nil || !validText(input.SourceURL, 2048, false) {
			return metadata.Movie{}, ErrInput
		}
	}
	for _, alias := range input.Aliases {
		if !validText(alias, 512, false) {
			return metadata.Movie{}, ErrInput
		}
	}
	info := metadata.Movie{Title: input.Title, OriginalTitle: input.OriginalTitle, Overview: input.Overview, Year: input.Year, ReleaseDate: input.ReleaseDate,
		Language: input.Language, OriginalLanguage: input.OriginalLanguage, Aliases: strings.Join(input.Aliases, "\n"), ExternalIDs: input.IDs, Source: source, SourceURL: input.SourceURL, Status: "matched"}
	for provider, id := range input.IDs {
		if !idPattern.MatchString(provider) || id == "" || !validText(id, 128, false) {
			return metadata.Movie{}, ErrInput
		}
		if provider == "tmdb" || provider == "bangumi" {
			number, err := strconv.ParseInt(id, 10, 64)
			if err != nil || number < 1 || number > 2147483647 {
				return metadata.Movie{}, ErrInput
			}
			if provider == "tmdb" {
				info.TMDBID = number
			} else {
				info.BangumiID = number
			}
		}
	}
	if source != "agent" && source != "nfo" {
		return metadata.Movie{}, ErrInput
	}
	return info, nil
}

func FromMovie(info metadata.Movie) Input {
	ids := map[string]string{}
	for k, v := range info.ExternalIDs {
		ids[k] = v
	}
	if info.TMDBID > 0 {
		ids["tmdb"] = strconv.FormatInt(info.TMDBID, 10)
	}
	if info.BangumiID > 0 {
		ids["bangumi"] = strconv.FormatInt(info.BangumiID, 10)
	}
	aliases := []string{}
	if info.Aliases != "" {
		aliases = strings.Split(info.Aliases, "\n")
	}
	return Input{Title: info.Title, OriginalTitle: info.OriginalTitle, Overview: info.Overview, Year: info.Year, ReleaseDate: info.ReleaseDate,
		Language: info.Language, OriginalLanguage: info.OriginalLanguage, Aliases: aliases, IDs: ids, SourceURL: info.SourceURL}
}
