package store

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"dinosaur/internal/media"
	"dinosaur/internal/playback"
)

func (s *Store) PlaybackProgress(ctx context.Context, userID, mediaID int64) (playback.Progress, error) {
	p := playback.Progress{MediaID: mediaID}
	err := s.DB.QueryRowContext(ctx, `SELECT p.position,p.watched,p.updated_at,m.duration FROM playback_progress p JOIN media_items m ON m.id=p.media_id WHERE p.user_id=? AND p.media_id=?`, userID, mediaID).Scan(&p.Position, &p.Watched, &p.UpdatedAt, &p.Duration)
	if errors.Is(err, sql.ErrNoRows) {
		return p, nil
	}
	return p, err
}
func (s *Store) SavePlaybackProgress(ctx context.Context, userID int64, p playback.Progress) error {
	_, err := s.DB.ExecContext(ctx, `INSERT INTO playback_progress(user_id,media_id,position,watched,updated_at) VALUES (?,?,?,?,?) ON CONFLICT(user_id,media_id) DO UPDATE SET position=excluded.position,watched=excluded.watched,updated_at=excluded.updated_at`, userID, p.MediaID, p.Position, p.Watched, time.Now().Unix())
	return err
}
func (s *Store) AddPlaybackProgress(ctx context.Context, userID int64, items []media.Item) error {
	if len(items) == 0 {
		return nil
	}
	placeholders := make([]string, len(items))
	args := make([]any, 1, len(items)+1)
	args[0] = userID
	positions := make(map[int64]int, len(items))
	for index, item := range items {
		placeholders[index] = "?"
		args = append(args, item.ID)
		positions[item.ID] = index
	}
	rows, err := s.DB.QueryContext(ctx, `SELECT media_id,position,watched,updated_at FROM playback_progress WHERE user_id=? AND media_id IN (`+strings.Join(placeholders, ",")+`)`, args...)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var id, updated int64
		var position float64
		var watched bool
		if err := rows.Scan(&id, &position, &watched, &updated); err != nil {
			return err
		}
		index := positions[id]
		items[index].Position = position
		items[index].Watched = watched
		items[index].LastPlayedAt = updated
	}
	return rows.Err()
}
func (s *Store) ContinueWatching(ctx context.Context, userID int64) ([]media.Item, error) {
	rows, err := s.DB.QueryContext(ctx, `SELECT `+mediaColumns+mediaFrom+` JOIN playback_progress p ON p.media_id=m.id WHERE p.user_id=? AND p.position>0 AND p.watched=0 AND m.missing=0 AND l.type='movie' ORDER BY p.updated_at DESC,p.media_id DESC LIMIT 12`, userID)
	if err != nil {
		return nil, err
	}
	result := make([]media.Item, 0)
	for rows.Next() {
		item, err := scanMedia(rows)
		if err != nil {
			rows.Close()
			return nil, err
		}
		result = append(result, item)
	}
	rowsErr := rows.Err()
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if rowsErr != nil {
		return nil, rowsErr
	}
	return result, s.AddPlaybackProgress(ctx, userID, result)
}
