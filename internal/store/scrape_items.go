package store

import (
	"context"
	"time"

	"dinosaur/internal/media"
)

func (s *Store) StartScrapeItems(ctx context.Context, jobID int64, items []media.Item) error {
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
		_, err := tx.ExecContext(ctx, `INSERT INTO scrape_job_items(job_id,source_media_id,media_id,title,filename,library_name,tmdb_id,bangumi_id,updated_at) VALUES (?,?,?,?,?,?,?,?,?)`,
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

func (s *Store) SaveScrapeItem(ctx context.Context, jobID, mediaID int64, status, message string) error {
	_, err := s.DB.ExecContext(ctx, `UPDATE scrape_job_items SET status=?,
 title=COALESCE(NULLIF((SELECT title FROM movie_metadata WHERE media_id=?),''),title),
 tmdb_id=COALESCE((SELECT tmdb_id FROM movie_metadata WHERE media_id=?),tmdb_id),
 bangumi_id=COALESCE((SELECT bangumi_id FROM movie_metadata WHERE media_id=?),bangumi_id),
 message=CASE WHEN ?='' THEN COALESCE((SELECT message FROM movie_metadata WHERE media_id=?),'') ELSE ? END,
 updated_at=? WHERE job_id=? AND source_media_id=?`, status, mediaID, mediaID, mediaID, message, mediaID, message, time.Now().Unix(), jobID, mediaID)
	return err
}
