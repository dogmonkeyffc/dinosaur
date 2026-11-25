package store

import (
	"context"
	"database/sql"
	"errors"
	"strconv"

	"dinosaur/internal/history"
)

func completedHistorySeries(ctx context.Context, tx *sql.Tx, userID, mediaID int64, e *history.Entry) (bool, error) {
	var seriesID int64
	err := tx.QueryRowContext(ctx, `SELECT series_id,work_id,title,year FROM history_episode_catalog WHERE media_id=?`, mediaID).
		Scan(&seriesID, &e.WorkID, &e.Title, &e.Year)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	filter, target := `ep.series_id=?`, seriesID
	e.SeriesID = &seriesID
	e.SourceID = "series:" + strconv.FormatInt(seriesID, 10)
	if e.WorkID != nil {
		filter, target = `ep.series_id IN (SELECT series_id FROM work_members WHERE work_id=?)`, *e.WorkID
		e.SeriesID = nil
		e.SourceID = "work:" + strconv.FormatInt(target, 10)
		// A work may combine series with different release years.
		e.Year = 0
	}
	var ready bool
	err = tx.QueryRowContext(ctx, `SELECT NOT EXISTS (
 SELECT 1 FROM tv_episodes ep JOIN media_items m ON m.id=ep.media_id
 LEFT JOIN playback_progress p ON p.media_id=m.id AND p.user_id=?
 WHERE m.missing=0 AND ep.season_number>0 AND `+filter+`
 GROUP BY ep.series_id,ep.season_number,ep.episode_number,ep.episode_end
 HAVING MAX(p.completed_at) IS NULL
 )`, userID, target).Scan(&ready)
	return ready, err
}
