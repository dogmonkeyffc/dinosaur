package store

import (
	"context"
	"database/sql"
	"errors"

	"dinosaur/internal/media"
	"dinosaur/internal/television"
)

func (s *Store) NextEpisode(ctx context.Context, item media.Item) (*media.Item, error) {
	if item.SeriesID == 0 {
		return nil, nil
	}
	next, err := scanMedia(s.DB.QueryRowContext(ctx, `SELECT `+mediaColumns+mediaFrom+` WHERE ep.series_id=? AND m.missing=0
 AND (ep.season_number>? OR (ep.season_number=? AND ep.episode_number>?))
 ORDER BY ep.season_number,ep.episode_number,m.id LIMIT 1`, item.SeriesID, item.SeasonNumber, item.SeasonNumber, item.EpisodeEnd))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &next, nil
}

func (s *Store) ContinueSeries(ctx context.Context, userID int64) ([]media.Item, error) {
	rows, err := s.DB.QueryContext(ctx, `SELECT e.series_id FROM tv_episodes e JOIN media_items m ON m.id=e.media_id
 JOIN playback_progress p ON p.media_id=m.id WHERE m.missing=0 AND p.user_id=? AND (p.position>0 OR p.watched=1)
 GROUP BY e.series_id ORDER BY max(p.updated_at) DESC,e.series_id DESC LIMIT 12`, userID)
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
	result := []media.Item{}
	for _, id := range ids {
		items, err := s.SeriesEpisodes(ctx, id, userID)
		if err != nil {
			return nil, err
		}
		if item := television.ResumeEpisode(items); item != nil && !item.Watched {
			result = append(result, *item)
		}
	}
	return result, nil
}
