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
 AND `+libraryVisible("m.library_id")+` AND (ep.season_number>? OR (ep.season_number=? AND ep.episode_number>?))
 ORDER BY ep.season_number,ep.episode_number,m.id LIMIT 1`, item.SeriesID, libraryViewer(ctx), item.SeasonNumber, item.SeasonNumber, item.EpisodeEnd))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &next, nil
}

func (s *Store) ContinueSeries(ctx context.Context, userID int64) ([]media.Item, error) {
	return s.continueSeries(ctx, userID, nil, false)
}

func (s *Store) continueSeries(ctx context.Context, userID int64, scope any, useSeriesTime bool) ([]media.Item, error) {
	rows, err := s.DB.QueryContext(ctx, `SELECT e.series_id,max(p.updated_at) FROM tv_episodes e JOIN media_items m ON m.id=e.media_id
 JOIN visible_playback_progress p ON p.media_id=m.id WHERE m.missing=0 AND p.user_id=? AND (p.position>0 OR p.watched=1)
 AND `+libraryVisible("m.library_id")+` AND `+resumeLibraryFilter+`
 AND EXISTS (SELECT 1 FROM tv_episodes available JOIN media_items entry ON entry.id=available.media_id
 LEFT JOIN visible_playback_progress saved ON saved.media_id=entry.id AND saved.user_id=?
 WHERE available.series_id=e.series_id AND entry.missing=0 AND coalesce(saved.watched,0)=0)
 GROUP BY e.series_id ORDER BY max(p.updated_at) DESC,e.series_id DESC LIMIT 12`, userID, libraryViewer(ctx), scope, scope, userID)
	if err != nil {
		return nil, err
	}
	type recentSeries struct{ id, updated int64 }
	var series []recentSeries
	for rows.Next() {
		var entry recentSeries
		if err := rows.Scan(&entry.id, &entry.updated); err != nil {
			rows.Close()
			return nil, err
		}
		series = append(series, entry)
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return nil, err
	}
	result := []media.Item{}
	for _, entry := range series {
		items, err := s.SeriesEpisodes(ctx, entry.id, userID)
		if err != nil {
			return nil, err
		}
		if item := television.ResumeEpisode(items); item != nil && !item.Watched {
			if useSeriesTime {
				item.LastPlayedAt = entry.updated
			}
			result = append(result, *item)
		}
	}
	return result, nil
}
