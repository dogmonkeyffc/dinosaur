package store

import (
	"context"
	"database/sql"

	"dinosaur/internal/television"
)

// WorkByID includes missing members in AllMembers for lossless administrative edits.
func (s *Store) WorkByID(ctx context.Context, id, userID int64) (television.WorkDetail, error) {
	result := television.WorkDetail{Members: []television.Series{}, AllMembers: []television.Series{}}
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return result, err
	}
	defer tx.Rollback()
	err = tx.QueryRowContext(ctx, `SELECT id,title,library_id FROM works WHERE id=?`, id).Scan(&result.Work.ID, &result.Work.Title, &result.Work.LibraryID)
	if err != nil {
		return result, err
	}
	rows, err := tx.QueryContext(ctx, `SELECT `+seriesColumns+seriesFrom+` JOIN work_members wm ON wm.series_id=s.id WHERE wm.work_id=? ORDER BY wm.position`, id)
	if err != nil {
		return result, err
	}
	for rows.Next() {
		item, err := scanSeries(rows)
		if err != nil {
			rows.Close()
			return result, err
		}
		result.AllMembers = append(result.AllMembers, item)
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return result, err
	}
	for i := range result.AllMembers {
		item := &result.AllMembers[i]
		err := tx.QueryRowContext(ctx, `SELECT count(*) FROM tv_episodes e JOIN media_items m ON m.id=e.media_id JOIN playback_progress p ON p.media_id=m.id WHERE e.series_id=? AND m.missing=0 AND p.user_id=? AND p.watched=1`, item.ID, userID).Scan(&item.WatchedCount)
		if err != nil {
			return result, err
		}
		if item.EpisodeCount == 0 {
			continue
		}
		result.Members = append(result.Members, *item)
		result.Work.EpisodeCount += item.EpisodeCount
		result.Work.SeasonCount += item.SeasonCount
		result.Work.WatchedCount += item.WatchedCount
		result.Work.MemberCount++
	}
	return result, tx.Commit()
}

func (s *Store) Works(ctx context.Context, libraryID, userID int64) ([]television.Work, error) {
	rows, err := s.DB.QueryContext(ctx, `SELECT w.id FROM works w WHERE (?=0 OR w.library_id=?) AND EXISTS(SELECT 1 FROM work_members wm JOIN tv_series s ON s.id=wm.series_id WHERE wm.work_id=w.id AND `+seriesVisible+`) ORDER BY w.title COLLATE NOCASE,w.id`, libraryID, libraryID)
	if err != nil {
		return nil, err
	}
	ids := []int64{}
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			rows.Close()
			return nil, err
		}
		ids = append(ids, id)
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return nil, err
	}
	result := []television.Work{}
	for _, id := range ids {
		detail, err := s.WorkByID(ctx, id, userID)
		if err == sql.ErrNoRows {
			continue
		}
		if err != nil {
			return nil, err
		}
		result = append(result, detail.Work)
	}
	return result, nil
}
