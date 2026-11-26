package scraping

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"dinosaur/internal/metadata"
)

type SeasonPosterRepository interface {
	MetadataSettings(context.Context) (metadata.Settings, error)
	PendingSeasonPosters(context.Context, int) ([]metadata.SeasonPosterTarget, error)
	SaveSeasonPoster(context.Context, int64, int, string, time.Time) error
}

type SeasonPosterProvider interface {
	SeasonPosterPaths(context.Context, metadata.Settings, int64) (map[int]string, error)
	CachePoster(context.Context, string, string) (string, error)
}

func StartSeasonPosters(repo SeasonPosterRepository, provider SeasonPosterProvider, dataDir string, logger *slog.Logger) func() {
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		defer close(done)
		ticker := time.NewTicker(time.Minute)
		defer ticker.Stop()
		for {
			if err := cacheSeasonPosters(ctx, repo, provider, dataDir); err != nil && ctx.Err() == nil {
				logger.Warn("season poster refresh failed", "error", err)
			}
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
			}
		}
	}()
	return func() {
		cancel()
		<-done
	}
}

func cacheSeasonPosters(ctx context.Context, repo SeasonPosterRepository, provider SeasonPosterProvider, dataDir string) error {
	settings, err := repo.MetadataSettings(ctx)
	if err != nil || settings.Credential == "" {
		return err
	}
	targets, err := repo.PendingSeasonPosters(ctx, 8)
	if err != nil {
		return err
	}
	for _, target := range targets {
		paths, err := provider.SeasonPosterPaths(ctx, settings, target.TMDBID)
		if errors.Is(err, metadata.ErrProviderAuth) || errors.Is(err, metadata.ErrProviderRateLimited) {
			return err
		}
		delay := 7 * 24 * time.Hour
		if err != nil && !errors.Is(err, metadata.ErrProviderNotFound) {
			delay = 15 * time.Minute
		}
		for _, number := range target.Numbers {
			if err := ctx.Err(); err != nil {
				return err
			}
			var key string
			retry := delay
			if poster := paths[number]; poster != "" {
				var downloadErr error
				key, downloadErr = provider.CachePoster(ctx, poster, dataDir)
				if downloadErr != nil {
					retry = 15 * time.Minute
				}
			}
			if err := repo.SaveSeasonPoster(ctx, target.TMDBID, number, key, time.Now().Add(retry)); err != nil {
				return err
			}
		}
	}
	return nil
}
