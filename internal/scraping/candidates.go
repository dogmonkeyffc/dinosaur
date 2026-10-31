package scraping

import (
	"context"
	"strings"

	"dinosaur/internal/media"
	"dinosaur/internal/metadata"
)

type SearchResult struct {
	Candidates []metadata.Candidate `json:"candidates"`
	Query      string               `json:"query"`
	Year       int                  `json:"year"`
}

func (m *Manager) Candidates(ctx context.Context, mediaID int64, query string, year *int) (SearchResult, error) {
	if err := m.validateScope(ctx, metadata.ScrapeRequest{MediaID: mediaID}); err != nil {
		return SearchResult{}, err
	}
	item, err := m.repo.MediaByID(ctx, mediaID)
	if err != nil {
		return SearchResult{}, err
	}
	name, releaseYear := media.ParseFilename(item.Filename)
	if query != "" {
		name = strings.TrimSpace(query)
	}
	if year != nil {
		releaseYear = *year
	}
	if name == "" || len(name) > 256 || releaseYear < 0 || releaseYear > 2100 {
		return SearchResult{}, ErrRequest
	}
	if query == "" && year == nil && item.Metadata != nil && len(item.Metadata.Candidates) > 0 {
		return SearchResult{Candidates: item.Metadata.Candidates, Query: name, Year: releaseYear}, nil
	}
	settings, err := m.repo.MetadataSettings(ctx)
	if err != nil {
		return SearchResult{}, err
	}
	candidates, err := m.search(ctx, settings, name, releaseYear, true)
	if err != nil {
		return SearchResult{}, err
	}
	candidates = metadata.TopCandidates(name, releaseYear, candidates)
	if query == "" && year == nil && item.Metadata != nil && item.Metadata.Status == "needs_review" {
		if err := m.repo.RecordScrapeReview(ctx, mediaID, candidates); err != nil {
			return SearchResult{}, err
		}
	}
	return SearchResult{Candidates: candidates, Query: name, Year: releaseYear}, err
}
