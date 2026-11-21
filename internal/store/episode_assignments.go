package store

import (
	"context"

	"dinosaur/internal/television"
)

func (s *Store) SaveEpisodeAssignments(ctx context.Context, libraryID int64, assignments []television.EpisodeAssignment) error {
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	for _, a := range assignments {
		result, err := tx.ExecContext(ctx, `
INSERT INTO episode_assignments(library_id,relative_path,series_folder,season_number,episode_number)
SELECT ?,?,?,?,?
WHERE NOT EXISTS(SELECT 1 FROM media_items WHERE library_id=? AND relative_path=?)
   OR EXISTS(SELECT 1 FROM episode_assignments WHERE library_id=? AND relative_path=?)
ON CONFLICT(library_id,relative_path) DO UPDATE SET episode_number=excluded.episode_number
WHERE episode_assignments.series_folder=excluded.series_folder
  AND episode_assignments.season_number=excluded.season_number
  AND episode_assignments.episode_number=excluded.episode_number`,
			libraryID, a.Path, a.SeriesFolder, a.Season, a.Episode, libraryID, a.Path, libraryID, a.Path)
		if err != nil {
			return err
		}
		count, err := result.RowsAffected()
		if err != nil {
			return err
		}
		if count == 0 {
			return television.ErrEpisodeAssigned
		}
	}
	return tx.Commit()
}

func (s *Store) EpisodeAssignments(ctx context.Context, libraryID int64) (map[string]television.EpisodeName, error) {
	rows, err := s.DB.QueryContext(ctx, `SELECT relative_path,series_folder,season_number,episode_number FROM episode_assignments WHERE library_id=?`, libraryID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := map[string]television.EpisodeName{}
	for rows.Next() {
		var a television.EpisodeAssignment
		if err := rows.Scan(&a.Path, &a.SeriesFolder, &a.Season, &a.Episode); err != nil {
			return nil, err
		}
		name, err := a.Name()
		if err != nil {
			return nil, err
		}
		result[a.Path] = name
	}
	return result, rows.Err()
}
