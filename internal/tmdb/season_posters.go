package tmdb

import (
	"context"
	"net/url"
	"regexp"
	"strconv"

	"dinosaur/internal/metadata"
)

var imdbSeriesID = regexp.MustCompile(`^tt[0-9]{7,12}$`)

func (c Television) FindSeasonPosterSeries(ctx context.Context, settings metadata.Settings, imdb string) (int64, error) {
	if !imdbSeriesID.MatchString(imdb) {
		return 0, metadata.ErrProviderNotFound
	}
	var response struct {
		Series []struct {
			ID int64 `json:"id"`
		} `json:"tv_results"`
	}
	if err := c.getJSON(ctx, settings, "/find/"+imdb, url.Values{"external_source": {"imdb_id"}}, &response); err != nil {
		return 0, err
	}
	if len(response.Series) != 1 || response.Series[0].ID <= 0 {
		return 0, metadata.ErrProviderNotFound
	}
	return response.Series[0].ID, nil
}

func (c Television) SeasonPosterPaths(ctx context.Context, settings metadata.Settings, seriesID int64) (map[int]string, error) {
	var response struct {
		ID      int64 `json:"id"`
		Seasons []struct {
			Number int    `json:"season_number"`
			Poster string `json:"poster_path"`
		} `json:"seasons"`
	}
	if err := c.getJSON(ctx, settings, "/tv/"+strconv.FormatInt(seriesID, 10), nil, &response); err != nil {
		return nil, err
	}
	if response.ID != seriesID {
		return nil, ErrUnavailable
	}
	paths := make(map[int]string, len(response.Seasons))
	for _, season := range response.Seasons {
		if posterPath.MatchString(season.Poster) {
			paths[season.Number] = season.Poster
		}
	}
	return paths, nil
}
