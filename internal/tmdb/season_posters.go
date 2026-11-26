package tmdb

import (
	"context"
	"strconv"

	"dinosaur/internal/metadata"
)

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
