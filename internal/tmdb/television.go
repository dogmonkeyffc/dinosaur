package tmdb

import (
	"context"
	"net/url"
	"strconv"
	"strings"
	"unicode"

	"dinosaur/internal/metadata"
)

// Television adapts the TV namespace to the shared background metadata worker.
type Television struct{ *Client }

type seriesResult struct {
	ID           int64  `json:"id"`
	Name         string `json:"name"`
	OriginalName string `json:"original_name"`
	Overview     string `json:"overview"`
	FirstAirDate string `json:"first_air_date"`
	PosterPath   string `json:"poster_path"`
}

func (c Television) Search(ctx context.Context, settings metadata.Settings, title string, year int) ([]metadata.Candidate, error) {
	query := url.Values{"query": {title}, "include_adult": {"false"}}
	if year > 0 {
		query.Set("first_air_date_year", strconv.Itoa(year))
	}
	var response struct {
		Results []seriesResult `json:"results"`
	}
	if err := c.getJSON(ctx, settings, "/search/tv", query, &response); err != nil {
		return nil, err
	}
	result := []metadata.Candidate{}
	for _, item := range response.Results {
		if item.ID > 0 {
			result = append(result, metadata.Candidate{ID: item.ID, Title: item.Name, OriginalTitle: item.OriginalName, Year: movieYear(item.FirstAirDate), Overview: item.Overview})
		}
	}
	return result, nil
}

func (c Television) Movie(ctx context.Context, settings metadata.Settings, id int64) (metadata.Movie, error) {
	var response struct {
		seriesResult
		Translations struct {
			Items []translation `json:"translations"`
		} `json:"translations"`
		Alternatives struct {
			Results []struct {
				Title string `json:"title"`
			} `json:"results"`
		} `json:"alternative_titles"`
	}
	err := c.getJSON(ctx, settings, "/tv/"+strconv.FormatInt(id, 10), url.Values{"append_to_response": {"translations,alternative_titles"}}, &response)
	if err != nil {
		return metadata.Movie{}, err
	}
	if response.ID != id || strings.TrimSpace(response.Name) == "" {
		return metadata.Movie{}, ErrUnavailable
	}
	info := metadata.Movie{TMDBID: id, Title: response.Name, OriginalTitle: response.OriginalName, Overview: response.Overview,
		Year: movieYear(response.FirstAirDate), ReleaseDate: response.FirstAirDate, PosterPath: response.PosterPath, Language: settings.Language, Status: "matched"}
	aliases := []string{response.Name, response.OriginalName}
	for i := range response.Translations.Items {
		item := &response.Translations.Items[i]
		item.Data.Title = item.Data.Name
		aliases = append(aliases, item.Data.Name)
	}
	for _, item := range response.Alternatives.Results {
		aliases = append(aliases, item.Title)
	}
	if strings.HasPrefix(settings.Language, "zh-") {
		localizeChinese(&info, settings.Language, response.Translations.Items)
		if !strings.ContainsFunc(info.Title, func(r rune) bool { return unicode.Is(unicode.Han, r) }) {
			info.Message = "no_chinese_title"
		}
	}
	info.Aliases = strings.Join(aliases, "\n")
	return info, nil
}
