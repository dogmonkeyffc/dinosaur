package store

import (
	"context"
	"time"

	"dinosaur/internal/metadata"
)

const scrapeJobColumns = `id,library_id,media_id,status,total,processed,matched,review,errors,skipped,message,created_at,ended_at,media_type`

func (s *Store) NewScrapeJob(ctx context.Context, request metadata.ScrapeRequest) (metadata.Job, error) {
	kind := request.MediaType
	if kind == "" {
		kind = "movie"
	}
	job := metadata.Job{MediaType: kind, LibraryID: request.LibraryID, MediaID: request.MediaID, Status: "queued", CreatedAt: time.Now().Unix()}
	result, err := s.DB.ExecContext(ctx, `INSERT INTO scrape_jobs(library_id,media_id,status,created_at,media_type) VALUES (?,?,?,?,?)`,
		job.LibraryID, job.MediaID, job.Status, job.CreatedAt, job.MediaType)
	if err != nil {
		return job, err
	}
	job.ID, err = result.LastInsertId()
	return job, err
}

func (s *Store) SaveScrapeJob(ctx context.Context, job metadata.Job) error {
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	_, err = tx.ExecContext(ctx, `UPDATE scrape_jobs SET status=?,total=?,processed=?,matched=?,review=?,errors=?,skipped=?,message=?,ended_at=? WHERE id=?`,
		job.Status, job.Total, job.Processed, job.Matched, job.Review, job.Errors, job.Skipped, job.Message, job.EndedAt, job.ID)
	if err != nil {
		return err
	}
	if job.Status != "queued" && job.Status != "running" {
		state := "not_processed"
		if job.Status == "interrupted" {
			state = "interrupted"
		}
		if _, err := tx.ExecContext(ctx, `UPDATE scrape_job_items SET status=?,message=?,updated_at=? WHERE job_id=? AND status='pending'`, state, job.Message, time.Now().Unix(), job.ID); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (s *Store) ScrapeJobs(ctx context.Context) ([]metadata.Job, error) {
	rows, err := s.DB.QueryContext(ctx, `SELECT `+scrapeJobColumns+` FROM scrape_jobs ORDER BY id DESC LIMIT 10`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	jobs := make([]metadata.Job, 0)
	for rows.Next() {
		job, err := scanScrapeJob(rows)
		if err != nil {
			return nil, err
		}
		jobs = append(jobs, job)
	}
	return jobs, rows.Err()
}

func (s *Store) InterruptScrapes(ctx context.Context) error {
	if _, err := s.DB.ExecContext(ctx, `UPDATE scrape_job_items SET status='interrupted',message='interrupted',updated_at=? WHERE status='pending' AND job_id IN (SELECT id FROM scrape_jobs WHERE status IN ('queued','running'))`, time.Now().Unix()); err != nil {
		return err
	}
	_, err := s.DB.ExecContext(ctx, `UPDATE scrape_jobs SET status='interrupted',message='interrupted',ended_at=? WHERE status IN ('queued','running')`, time.Now().Unix())
	return err
}
