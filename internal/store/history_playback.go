package store

import (
	"context"
	"database/sql"
	"strconv"
	"time"

	"dinosaur/internal/history"
	"dinosaur/internal/playback"
)

func (s *Store) saveProgressAndHistory(ctx context.Context, userID int64, p playback.Progress) error {
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var watched bool
	err = tx.QueryRowContext(ctx, `SELECT watched FROM playback_progress WHERE user_id=? AND media_id=?`, userID, p.MediaID).Scan(&watched)
	if err != nil && err != sql.ErrNoRows {
		return err
	}
	now := time.Now()
	_, err = tx.ExecContext(ctx, `INSERT INTO playback_progress(user_id,media_id,position,watched,updated_at) VALUES (?,?,?,?,?) ON CONFLICT(user_id,media_id) DO UPDATE SET position=excluded.position,watched=excluded.watched,updated_at=excluded.updated_at`, userID, p.MediaID, p.Position, p.Watched, now.Unix())
	if err != nil {
		return err
	}
	if p.Watched && !watched {
		e := history.Entry{MediaID: &p.MediaID, Source: "local", WatchedOn: history.Day(now)}
		e.SourceID = strconv.FormatInt(p.MediaID, 10) + ":" + e.WatchedOn
		err = tx.QueryRowContext(ctx, `SELECT `+displayTitle+`,COALESCE(NULLIF(md.year,0),m.year)`+mediaFrom+` WHERE m.id=?`, p.MediaID).Scan(&e.Title, &e.Year)
		if err != nil {
			return err
		}
		if _, err = tx.ExecContext(ctx, insertHistory+` ON CONFLICT(user_id,source,source_id,watched_on) DO NOTHING`, historyArgs(userID, e)...); err != nil {
			return err
		}
	}
	return tx.Commit()
}
