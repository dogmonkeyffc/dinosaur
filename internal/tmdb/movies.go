package tmdb

import (
	"context"
	"net/url"
	"strconv"
	"strings"
	"unicode"

	"dinosaur/internal/metadata"
)

type movieResult struct {
	OriginalLanguage string `json:"original_language"`
	ID               int64  `json:"id"`
	Title            string `json:"title"`
	OriginalTitle    string `json:"original_title"`
	Overview         string `json:"overview"`
	ReleaseDate      string `json:"release_date"`
	PosterPath       string `json:"poster_path"`
}

func movieYear(date string) int {
	if len(date) < 4 {
		return 0
	}
	year, _ := strconv.Atoi(date[:4])
	return year
}

func (c *Client) Search(ctx context.Context, settings metadata.Settings, title string, year int) ([]metadata.Candidate, error) {
	query := url.Values{"query": {title}, "include_adult": {"false"}}
	if year > 0 {
		query.Set("primary_release_year", strconv.Itoa(year))
	}
	var result struct {
		Results []movieResult `json:"results"`
	}
	if err := c.getJSON(ctx, settings, "/search/movie", query, &result); err != nil {
		return nil, err
	}
	items := make([]metadata.Candidate, 0)
	for _, item := range result.Results {
		if item.ID <= 0 {
			continue
		}
		items = append(items, metadata.Candidate{ID: item.ID, Title: item.Title, OriginalTitle: item.OriginalTitle,
			Year: movieYear(item.ReleaseDate), Overview: item.Overview})
	}
	return items, nil
}

func (c *Client) Movie(ctx context.Context, settings metadata.Settings, id int64) (metadata.Movie, error) {
	var result struct {
		movieResult
		Translations struct {
			Items []translation `json:"translations"`
		} `json:"translations"`
		Alternatives struct {
			Titles []struct {
				Title string `json:"title"`
			} `json:"titles"`
		} `json:"alternative_titles"`
	}
	query := url.Values{"append_to_response": {"translations,alternative_titles"}}
	if err := c.getJSON(ctx, settings, "/movie/"+strconv.FormatInt(id, 10), query, &result); err != nil {
		return metadata.Movie{}, err
	}
	if result.ID != id || strings.TrimSpace(result.Title) == "" {
		return metadata.Movie{}, ErrUnavailable
	}
	movie := metadata.Movie{TMDBID: id, Title: result.Title, OriginalTitle: result.OriginalTitle,
		Overview: result.Overview, Year: movieYear(result.ReleaseDate), ReleaseDate: result.ReleaseDate,
		PosterPath: result.PosterPath, OriginalLanguage: result.OriginalLanguage, Language: settings.Language, Status: "matched"}
	if strings.HasPrefix(settings.Language, "zh-") {
		localizeChinese(&movie, settings.Language, result.Translations.Items)
		if !strings.ContainsFunc(movie.Title, func(char rune) bool { return unicode.Is(unicode.Han, char) }) {
			movie.Message = "no_chinese_title"
		}
	}
	aliases := []string{result.Title, result.OriginalTitle}
	for _, item := range result.Alternatives.Titles {
		aliases = append(aliases, item.Title)
	}
	for _, item := range result.Translations.Items {
		aliases = append(aliases, item.Data.Title)
	}
	movie.Aliases = strings.Join(aliases, "\n")
	return movie, nil
}
