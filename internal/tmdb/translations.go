package tmdb

import (
	"strings"

	"dinosaur/internal/metadata"
)

type translation struct {
	Language string `json:"iso_639_1"`
	Country  string `json:"iso_3166_1"`
	Data     struct {
		Title    string `json:"title"`
		Name     string `json:"name"`
		Overview string `json:"overview"`
	} `json:"data"`
}

func localizeChinese(movie *metadata.Movie, language string, translations []translation) {
	countries := []string{strings.TrimPrefix(language, "zh-"), "CN", "TW", "HK"}
	titleFound, overviewFound := false, false
	for _, country := range countries {
		for _, item := range translations {
			if item.Language != "zh" || item.Country != country {
				continue
			}
			if !titleFound && strings.TrimSpace(item.Data.Title) != "" {
				movie.Title = item.Data.Title
				movie.Language = "zh-" + country
				titleFound = true
			}
			if !overviewFound && strings.TrimSpace(item.Data.Overview) != "" {
				movie.Overview = item.Data.Overview
				overviewFound = true
			}
		}
	}
}
