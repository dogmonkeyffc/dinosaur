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
	var record bool
	var kind string
	err = tx.QueryRowContext(ctx, `SELECT l.type,COALESCE(h.record_history,1)
 FROM media_items m JOIN libraries l ON l.id=m.library_id
 LEFT JOIN history_library_settings h ON h.library_id=l.id AND h.user_id=? WHERE m.id=?`, userID, p.MediaID).Scan(&kind, &record)
	if err != nil {
		return err
	}
	if !record {
		return tx.Commit()
	}
	now := time.Now()
	var completedAt *int64
	if p.Completed {
		stamp := now.Unix()
		completedAt = &stamp
	}
	_, err = tx.ExecContext(ctx, `INSERT INTO playback_progress(user_id,media_id,position,watched,updated_at,completed_at)
 VALUES (?,?,?,?,?,?) ON CONFLICT(user_id,media_id) DO UPDATE SET position=excluded.position,watched=excluded.watched,
 updated_at=excluded.updated_at,completed_at=COALESCE(playback_progress.completed_at,excluded.completed_at)`,
		userID, p.MediaID, p.Position, p.Watched, now.Unix(), completedAt)
	if err != nil {
		return err
	}
	if p.Completed {
		if err := addCompletionHistory(ctx, tx, userID, p.MediaID, kind, now); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func addCompletionHistory(ctx context.Context, tx *sql.Tx, userID, mediaID int64, kind string, now time.Time) error {
	e := history.Entry{Source: "local", WatchedOn: history.Day(now)}
	if kind == "tv" {
		ready, err := completedHistorySeries(ctx, tx, userID, mediaID, &e)
		if err != nil || !ready {
			return err
		}
	} else {
		e.MediaID = &mediaID
		e.SourceID = strconv.FormatInt(mediaID, 10) + ":" + e.WatchedOn
		err := tx.QueryRowContext(ctx, `SELECT `+displayTitle+`,COALESCE(NULLIF(md.year,0),m.year)`+mediaFrom+` WHERE m.id=?`, mediaID).Scan(&e.Title, &e.Year)
		if err != nil {
			return err
		}
	}
	_, err := tx.ExecContext(ctx, insertHistory+` ON CONFLICT DO NOTHING`, historyArgs(userID, e)...)
	return err
}
