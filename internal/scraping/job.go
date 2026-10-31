package scraping

import (
	"context"
	"errors"
	"time"

	"dinosaur/internal/media"
	"dinosaur/internal/metadata"
)

func (m *Manager) run(job metadata.Job, request metadata.ScrapeRequest, settings metadata.Settings, items []media.Item) {
	job.Status = "running"
	if err := m.repo.SaveScrapeJob(m.ctx, job); err != nil {
		job.Status, job.Message = "failed", "storage_error"
	} else {
		for _, item := range items {
			if request.TMDBID == 0 && request.BangumiID == 0 {
				m.runManualJobs()
			}
			if m.ctx.Err() != nil {
				break
			}
			ctx, cancel := context.WithTimeout(m.ctx, 40*time.Second)
			outcome, err := m.scrape(ctx, settings, request, item)
			cancel()
			if m.ctx.Err() != nil {
				break
			}
			state, message := outcome, ""
			if state == "review" {
				state = "needs_review"
			}
			if err != nil {
				message = errorMessage(err)
			}
			if saveErr := m.repo.SaveScrapeItem(m.ctx, job.ID, item.ID, state, message); saveErr != nil {
				job.Status, job.Message = "failed", "storage_error"
				break
			}
			job.Processed++
			switch outcome {
			case "matched":
				job.Matched++
			case "review":
				job.Review++
			case "skipped":
				job.Skipped++
			default:
				job.Errors++
			}
			if err != nil {
				job.Message = errorMessage(err)
			}
			if saveErr := m.repo.SaveScrapeJob(m.ctx, job); saveErr != nil {
				job.Status, job.Message = "failed", "storage_error"
				break
			}
			if errors.Is(err, metadata.ErrProviderAuth) || errors.Is(err, metadata.ErrProviderRateLimited) {
				job.Status = "failed"
				break
			}
			timer := time.NewTimer(300 * time.Millisecond)
			select {
			case <-m.ctx.Done():
			case <-timer.C:
			}
			timer.Stop()
		}
	}
	if job.Status == "running" {
		job.Status = "completed"
	}
	if m.ctx.Err() != nil {
		job.Status, job.Message = "interrupted", "interrupted"
	}
	job.EndedAt = time.Now().Unix()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = m.repo.SaveScrapeJob(ctx, job)
}

func errorMessage(err error) string {
	switch {
	case errors.Is(err, metadata.ErrConflict):
		return "metadata_conflict"
	case errors.Is(err, metadata.ErrProviderAuth):
		return "invalid_credential"
	case errors.Is(err, metadata.ErrProviderRateLimited):
		return "rate_limited"
	case errors.Is(err, metadata.ErrProviderNotFound):
		return "not_found"
	default:
		return "provider_unavailable"
	}
}
