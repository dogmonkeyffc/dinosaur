package store

import (
	"context"

	"dinosaur/internal/metadata"
)

func scanScrapeJob(row interface{ Scan(...any) error }) (metadata.Job, error) {
	var job metadata.Job
	err := row.Scan(&job.ID, &job.LibraryID, &job.MediaID, &job.Status, &job.Total, &job.Processed, &job.Matched,
		&job.Review, &job.Errors, &job.Skipped, &job.Message, &job.CreatedAt, &job.EndedAt, &job.MediaType)
	return job, err
}

func (s *Store) ScrapeJobPage(ctx context.Context, filter metadata.ReportFilter) (metadata.JobPage, error) {
	page := metadata.JobPage{Jobs: []metadata.Job{}, Page: filter.Page, PageSize: filter.PageSize}
	where := ` WHERE 1=1`
	args := []any{}
	if filter.MediaType != "" {
		where += ` AND media_type=?`
		args = append(args, filter.MediaType)
	}
	if filter.Status != "" {
		where += ` AND status=?`
		args = append(args, filter.Status)
	}
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return page, err
	}
	defer tx.Rollback()
	if err := tx.QueryRowContext(ctx, `SELECT count(*) FROM scrape_jobs`+where, args...).Scan(&page.Total); err != nil {
		return page, err
	}
	args = append(args, filter.PageSize, (filter.Page-1)*filter.PageSize)
	rows, err := tx.QueryContext(ctx, `SELECT `+scrapeJobColumns+` FROM scrape_jobs`+where+` ORDER BY id DESC LIMIT ? OFFSET ?`, args...)
	if err != nil {
		return page, err
	}
	for rows.Next() {
		job, err := scanScrapeJob(rows)
		if err != nil {
			rows.Close()
			return page, err
		}
		page.Jobs = append(page.Jobs, job)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return page, err
	}
	rows.Close()
	return page, tx.Commit()
}

func (s *Store) ScrapeJobDetail(ctx context.Context, id int64, filter metadata.ReportFilter) (metadata.JobDetail, error) {
	result := metadata.JobDetail{Items: []metadata.JobItem{}, Page: filter.Page, PageSize: filter.PageSize}
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return result, err
	}
	defer tx.Rollback()
	result.Job, err = scanScrapeJob(tx.QueryRowContext(ctx, `SELECT `+scrapeJobColumns+` FROM scrape_jobs WHERE id=?`, id))
	if err != nil {
		return result, err
	}
	if err := tx.QueryRowContext(ctx, `SELECT count(*) FROM scrape_job_items WHERE job_id=?`, id).Scan(&result.Total); err != nil {
		return result, err
	}
	rows, err := tx.QueryContext(ctx, `SELECT id,COALESCE(media_id,0),title,filename,library_name,status,message,tmdb_id,updated_at,COALESCE(series_id,0) FROM scrape_job_items WHERE job_id=? ORDER BY id LIMIT ? OFFSET ?`, id, filter.PageSize, (filter.Page-1)*filter.PageSize)
	if err != nil {
		return result, err
	}
	for rows.Next() {
		var item metadata.JobItem
		if err := rows.Scan(&item.ID, &item.MediaID, &item.Title, &item.Filename, &item.Library, &item.Status, &item.Message, &item.TMDBID, &item.UpdatedAt, &item.SeriesID); err != nil {
			rows.Close()
			return result, err
		}
		result.Items = append(result.Items, item)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return result, err
	}
	rows.Close()
	return result, tx.Commit()
}
