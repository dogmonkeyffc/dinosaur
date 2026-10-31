package store

import (
	"context"
	"encoding/json"
	"strings"

	"dinosaur/internal/media"
	"dinosaur/internal/metadata"
	"dinosaur/internal/television"
)

const seriesTitle = `COALESCE(NULLIF(json_extract(s.metadata_json,'$.title'),''),s.title)`
const seriesVisible = ` EXISTS(SELECT 1 FROM tv_episodes e JOIN media_items m ON m.id=e.media_id WHERE e.series_id=s.id AND m.missing=0)`
const seriesColumns = `s.id,s.library_id,l.name,s.metadata_version,s.aliases,` + seriesTitle + `,
 COALESCE(NULLIF(json_extract(s.metadata_json,'$.year'),0),s.year),s.metadata_json,
 COALESCE(NULLIF(s.poster_key,''),(SELECT m.poster_key FROM tv_episodes e JOIN media_items m ON m.id=e.media_id WHERE e.series_id=s.id AND m.missing=0 AND m.poster_key<>'' ORDER BY e.season_number,e.episode_number,m.id LIMIT 1),''),
 (SELECT count(*) FROM tv_episodes e JOIN media_items m ON m.id=e.media_id WHERE e.series_id=s.id AND m.missing=0),
 (SELECT count(DISTINCT e.season_number) FROM tv_episodes e JOIN media_items m ON m.id=e.media_id WHERE e.series_id=s.id AND m.missing=0)`
const seriesFrom = ` FROM tv_series s JOIN libraries l ON l.id=s.library_id`

func scanSeries(row interface{ Scan(...any) error }) (television.Series, error) {
	var item television.Series
	var raw, aliases string
	var version int64
	err := row.Scan(&item.ID, &item.LibraryID, &item.LibraryName, &version, &aliases, &item.Title, &item.Year, &raw, &item.PosterKey, &item.EpisodeCount, &item.SeasonCount)
	if err != nil {
		return item, err
	}
	var info metadata.Movie
	if err := json.Unmarshal([]byte(raw), &info); err != nil {
		return item, err
	}
	info.Version, info.Aliases, info.PosterKey = version, aliases, item.PosterKey
	if info.Status != "" {
		item.Metadata = &info
	}
	return item, nil
}

func (s *Store) SeriesByID(ctx context.Context, id int64) (television.Series, error) {
	return scanSeries(s.DB.QueryRowContext(ctx, `SELECT `+seriesColumns+seriesFrom+` WHERE s.id=? AND `+seriesVisible, id))
}

func (s *Store) SeriesPage(ctx context.Context, filter media.Filter, userID int64) (television.Page, error) {
	page := television.Page{Items: []television.Series{}, Page: filter.Page, PageSize: filter.PageSize}
	where := ` WHERE ` + seriesVisible
	args := []any{}
	if filter.LibraryID > 0 {
		where += ` AND s.library_id=?`
		args = append(args, filter.LibraryID)
	}
	if filter.Query != "" {
		where += ` AND (` + seriesTitle + ` LIKE ? ESCAPE '\' OR s.title LIKE ? ESCAPE '\' OR s.aliases LIKE ? ESCAPE '\')`
		query := "%" + strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`).Replace(filter.Query) + "%"
		args = append(args, query, query, query)
	}
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return page, err
	}
	defer tx.Rollback()
	if err := tx.QueryRowContext(ctx, `SELECT count(*)`+seriesFrom+where, args...).Scan(&page.Total); err != nil {
		return page, err
	}
	args = append(args, filter.PageSize, (filter.Page-1)*filter.PageSize)
	rows, err := tx.QueryContext(ctx, `SELECT `+seriesColumns+seriesFrom+where+` ORDER BY `+seriesTitle+` COLLATE NOCASE,s.id LIMIT ? OFFSET ?`, args...)
	if err != nil {
		return page, err
	}
	for rows.Next() {
		item, err := scanSeries(rows)
		if err != nil {
			rows.Close()
			return page, err
		}
		page.Items = append(page.Items, item)
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return page, err
	}
	for i := range page.Items {
		err := tx.QueryRowContext(ctx, `SELECT count(*) FROM tv_episodes e JOIN media_items m ON m.id=e.media_id
   JOIN playback_progress p ON p.media_id=m.id WHERE e.series_id=? AND m.missing=0 AND p.user_id=? AND p.watched=1`, page.Items[i].ID, userID).Scan(&page.Items[i].WatchedCount)
		if err != nil {
			return page, err
		}
	}
	return page, tx.Commit()
}

func (s *Store) SeriesEpisodes(ctx context.Context, id, userID int64) ([]media.Item, error) {
	rows, err := s.DB.QueryContext(ctx, `SELECT `+mediaColumns+mediaFrom+` WHERE ep.series_id=? AND m.missing=0 ORDER BY ep.season_number,ep.episode_number,m.id`, id)
	if err != nil {
		return nil, err
	}
	items := []media.Item{}
	for rows.Next() {
		item, err := scanMedia(rows)
		if err != nil {
			rows.Close()
			return nil, err
		}
		items = append(items, item)
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return nil, err
	}
	return items, s.AddPlaybackProgress(ctx, userID, items)
}
