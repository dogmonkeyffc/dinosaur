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
	result, err := s.DB.ExecContext(ctx, `INSERT INTO history_library_settings(user_id,library_id,record_history,updated_at)
 SELECT ?,l.id,?,? FROM libraries l WHERE l.id=? AND `+libraryVisible("l.id")+`
 ON CONFLICT(user_id,library_id) DO UPDATE SET record_history=excluded.record_history,updated_at=excluded.updated_at`,
		userID, record, time.Now().Unix(), libraryID, libraryViewer(ctx))
	if err != nil {
		return err
	}
	return historyAffected(result)
}
