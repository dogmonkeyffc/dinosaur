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
	if current.Metadata.Managed() && request.Refresh {
		return m.failure(ctx, item.ID, metadata.ErrConflict)
	}
	if current.Metadata.Matched() && !request.Refresh {
		return "skipped", nil
	}
	source, id := "tmdb", request.TMDBID
	if request.BangumiID > 0 {
		source, id = "bangumi", request.BangumiID
	}
	if id == 0 && request.Refresh {
		source, id = current.Metadata.Identity()
	}
	automatic := id == 0
	query, year := media.ParseFilename(item.Filename)
	if automatic {
		candidates, err := m.search(ctx, settings, query, year, false)
		if err != nil {
			return m.failure(ctx, item.ID, err)
		}
		candidate, matched := metadata.AutoCandidate(query, year, candidates)
		if !matched {
			return "review", m.repo.RecordScrapeReview(ctx, item.ID, metadata.TopCandidates(query, year, candidates))
		}
		source, id = candidate.Provider, candidate.ID
	}
	provider := m.source(source)
	if provider == nil {
		return m.failure(ctx, item.ID, metadata.ErrProviderUnavailable)
	}
	movie, err := provider.Movie(ctx, providerSettings(settings, source), id)
	if err != nil && automatic && source != "bangumi" && m.fallback != nil {
		candidates, fallbackErr := m.secondary(ctx, settings, query, year)
		if fallbackErr == nil {
			if selected, matched := metadata.AutoCandidate(query, year, candidates); matched {
				source, id, provider = "bangumi", selected.ID, m.fallback
				movie, err = provider.Movie(ctx, providerSettings(settings, source), id)
			} else if len(candidates) > 0 {
				return "review", m.repo.RecordScrapeReview(ctx, item.ID, metadata.TopCandidates(query, year, candidates))
			}
		}
	}
	if err != nil {
		return m.failure(ctx, item.ID, err)
	}
	if movie.PosterPath != "" {
		key, err := provider.CachePoster(ctx, movie.PosterPath, m.dataDir)
		if err != nil {
			movie.Message = strings.Trim(movie.Message+",poster_unavailable", ",")
			previousSource, previousID := current.Metadata.Identity()
			if source == "" {
				source = "tmdb"
			}
			if previousSource == source && previousID == id {
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

func providerSettings(settings metadata.Settings, source string) metadata.Settings {
	if source == "bangumi" {
		return metadata.Settings{Language: settings.Language}
	}
	return settings
}

func (m *Manager) failure(ctx context.Context, id int64, cause error) (string, error) {
	if err := m.repo.RecordScrapeFailure(ctx, id, "failed", errorMessage(cause)); err != nil {
		return "failed", err
	}
	return "failed", cause
}
