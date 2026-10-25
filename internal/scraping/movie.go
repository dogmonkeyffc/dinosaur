package scraping

import (
	"context"
	"strings"

	"dinosaur/internal/media"
	"dinosaur/internal/metadata"
)

func (m *Manager) scrape(ctx context.Context, settings metadata.Settings, request metadata.ScrapeRequest, item media.Item) (string, error) {
	current, err := m.repo.MediaByID(ctx, item.ID)
	if err != nil {
		return "failed", err
	}
	if current.Metadata != nil && current.Metadata.TMDBID > 0 && !request.Refresh {
		return "skipped", nil
	}
	id := request.TMDBID
	if id == 0 && request.Refresh && current.Metadata != nil {
		id = current.Metadata.TMDBID
	}
	if id == 0 {
		query, year := media.ParseFilename(item.Filename)
		candidates, err := m.provider.Search(ctx, settings, query, year)
		if err != nil {
			return m.failure(ctx, item.ID, err)
		}
		if len(candidates) == 0 && year > 0 {
			candidates, err = m.provider.Search(ctx, settings, query, 0)
			if err != nil {
				return m.failure(ctx, item.ID, err)
			}
		}
		var matched bool
		id, matched = metadata.AutoMatch(query, year, candidates)
		if !matched {
			err := m.repo.RecordScrapeReview(ctx, item.ID, metadata.TopCandidates(query, year, candidates))
			return "review", err
		}
	}
	movie, err := m.provider.Movie(ctx, settings, id)
	if err != nil {
		return m.failure(ctx, item.ID, err)
	}
	if movie.PosterPath != "" {
		key, err := m.provider.CachePoster(ctx, movie.PosterPath, m.dataDir)
		if err != nil {
			movie.Message = strings.Trim(movie.Message+",poster_unavailable", ",")
			if current.Metadata != nil && current.Metadata.TMDBID == id {
				movie.PosterKey = current.PosterKey
			}
		} else {
			movie.PosterKey = key
		}
	}
	if err := m.repo.SaveMovieMetadata(ctx, item.ID, movie); err != nil {
		return "failed", err
	}
	return "matched", nil
}

func (m *Manager) failure(ctx context.Context, id int64, cause error) (string, error) {
	if err := m.repo.RecordScrapeFailure(ctx, id, "failed", errorMessage(cause)); err != nil {
		return "failed", err
	}
	return "failed", cause
}
