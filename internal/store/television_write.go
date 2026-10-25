package store

import (
	"context"

	"dinosaur/internal/media"
	"dinosaur/internal/television"
)

// SaveEpisode commits file information and its series association together.
func (s *Store) SaveEpisode(ctx context.Context, item media.Item, episode television.EpisodeName, jobID int64) error {
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var seriesID int64
	err = tx.QueryRowContext(ctx, `INSERT INTO tv_series(library_id,source_key,title,year) VALUES (?,?,?,?)
 ON CONFLICT(library_id,source_key) DO UPDATE SET title=excluded.title,year=excluded.year RETURNING id`,
		item.LibraryID, episode.Key, episode.Series, episode.Year).Scan(&seriesID)
	if err != nil {
		return err
	}
	if err := saveMedia(ctx, tx, item, jobID); err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, `INSERT INTO tv_episodes(media_id,series_id,season_number,episode_number,episode_end)
 SELECT id,?,?,?,? FROM media_items WHERE library_id=? AND relative_path=?
 ON CONFLICT(media_id) DO UPDATE SET series_id=excluded.series_id,season_number=excluded.season_number,
 episode_number=excluded.episode_number,episode_end=excluded.episode_end`,
		seriesID, episode.Season, episode.Episode, episode.End, item.LibraryID, item.RelativePath)
	if err != nil {
		return err
	}
	return tx.Commit()
}
