package store

import (
	"context"
	"time"

	"dinosaur/internal/media"
)

func (s TelevisionRepository) StartScrapeItems(ctx context.Context, jobID int64, items []media.Item) error {
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	for _, item := range items {
		var tmdbID, bangumiID int64
		if item.Metadata != nil {
			tmdbID, bangumiID = item.Metadata.TMDBID, item.Metadata.BangumiID
		}
		_, err := tx.ExecContext(ctx, `INSERT INTO scrape_job_items(job_id,source_media_id,series_id,title,filename,library_name,tmdb_id,bangumi_id,updated_at) VALUES (?,?,?,?,?,?,?,?,?)`,
			jobID, item.ID, item.ID, item.Title, item.Filename, item.LibraryName, tmdbID, bangumiID, time.Now().Unix())
		if err != nil {
			return err
		}
	}
	if _, err := tx.ExecContext(ctx, `UPDATE scrape_jobs SET total=? WHERE id=?`, len(items), jobID); err != nil {
		return err
	}
	return tx.Commit()
}

func (s TelevisionRepository) SaveScrapeItem(ctx context.Context, jobID, seriesID int64, status, message string) error {
	_, err := s.DB.ExecContext(ctx, `UPDATE scrape_job_items SET status=?,
 title=COALESCE(NULLIF((SELECT json_extract(metadata_json,'$.title') FROM tv_series WHERE id=?),''),title),
 tmdb_id=COALESCE((SELECT json_extract(metadata_json,'$.tmdb_id') FROM tv_series WHERE id=?),tmdb_id),
 bangumi_id=COALESCE((SELECT COALESCE(json_extract(metadata_json,'$.bangumi_id'),0) FROM tv_series WHERE id=?),bangumi_id),
 message=CASE WHEN ?='' THEN COALESCE((SELECT json_extract(metadata_json,'$.message') FROM tv_series WHERE id=?),'') ELSE ? END,
 updated_at=? WHERE job_id=? AND source_media_id=?`, status, seriesID, seriesID, seriesID, message, seriesID, message, time.Now().Unix(), jobID, seriesID)
	return err
}
