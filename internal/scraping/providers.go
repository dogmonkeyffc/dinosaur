package scraping

import (
	"context"
	"errors"

	"dinosaur/internal/metadata"
)

// WithFallback configures the secondary source before the manager is started.
func (m *Manager) WithFallback(provider Provider) *Manager {
	m.fallback = provider
	return m
}

func searchSource(ctx context.Context, provider Provider, settings metadata.Settings, query string, year int) ([]metadata.Candidate, error) {
	items, err := provider.Search(ctx, settings, query, year)
	if err == nil && len(items) == 0 && year > 0 {
		items, err = provider.Search(ctx, settings, query, 0)
	}
	return items, err
}

func (m *Manager) secondary(ctx context.Context, settings metadata.Settings, query string, year int) ([]metadata.Candidate, error) {
	if m.fallback == nil {
		return nil, nil
	}
	items, err := searchSource(ctx, m.fallback, metadata.Settings{Language: settings.Language}, query, year)
	items = append([]metadata.Candidate(nil), items...)
	for i := range items {
		items[i].Provider = "bangumi"
	}
	return items, err
}

func (m *Manager) search(ctx context.Context, settings metadata.Settings, query string, year int, all bool) ([]metadata.Candidate, error) {
	var items []metadata.Candidate
	var err error
	if settings.Credential != "" {
		items, err = searchSource(ctx, m.provider, settings, query, year)
	} else {
		err = metadata.ErrMissingCredential
	}
	if m.fallback == nil {
		return items, err
	}
	if _, matched := metadata.AutoCandidate(query, year, items); err == nil && matched && !all {
		return items, nil
	}
	fallback, fallbackErr := m.secondary(ctx, settings, query, year)
	if fallbackErr != nil {
		if err == nil && len(items) > 0 {
			return items, nil
		}
		return nil, fallbackErr
	}
	if err != nil && !errors.Is(err, metadata.ErrMissingCredential) && len(fallback) == 0 {
		return nil, err
	}
	return append(items, fallback...), nil
}

func (m *Manager) source(name string) Provider {
	if name == "bangumi" {
		return m.fallback
	}
	return m.provider
}
