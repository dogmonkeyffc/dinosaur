package store

import (
	"context"
	"time"

	"dinosaur/internal/history"
)

func (s *Store) HistoryLibrarySettings(ctx context.Context, userID int64) ([]history.LibrarySetting, error) {
	rows, err := s.DB.QueryContext(ctx, `SELECT l.id,l.name,l.type,COALESCE(p.record_history,1)
 FROM libraries l LEFT JOIN history_library_settings p ON p.library_id=l.id AND p.user_id=?
 WHERE `+libraryVisible("l.id")+` ORDER BY l.id`, userID, libraryViewer(ctx))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []history.LibrarySetting{}
	for rows.Next() {
		var item history.LibrarySetting
		if err := rows.Scan(&item.ID, &item.Name, &item.Type, &item.RecordHistory); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *Store) SetHistoryLibrarySetting(ctx context.Context, userID, libraryID int64, record bool) error {
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	result, err := tx.ExecContext(ctx, `INSERT INTO history_library_settings(user_id,library_id,record_history,updated_at)
 SELECT ?,l.id,?,? FROM libraries l WHERE l.id=? AND `+libraryVisible("l.id")+`
 ON CONFLICT(user_id,library_id) DO UPDATE SET record_history=excluded.record_history,updated_at=excluded.updated_at`,
		userID, record, time.Now().Unix(), libraryID, libraryViewer(ctx))
	if err != nil {
		return err
	}
	if err := historyAffected(result); err != nil {
		return err
	}
	if !record {
		if _, err := tx.ExecContext(ctx, `DELETE FROM playback_progress WHERE user_id=? AND media_id IN (SELECT id FROM media_items WHERE library_id=?)`, userID, libraryID); err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, `DELETE FROM watch_history WHERE user_id=? AND source='local' AND (
 media_id IN (SELECT id FROM media_items WHERE library_id=?) OR
 series_id IN (SELECT id FROM tv_series WHERE library_id=?) OR
 work_id IN (SELECT id FROM works WHERE library_id=?))`, userID, libraryID, libraryID, libraryID); err != nil {
			return err
		}
	}
	return tx.Commit()
}
