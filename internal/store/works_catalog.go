package store

import (
	"context"
	"strings"

	"dinosaur/internal/media"
	"dinosaur/internal/television"
)

// Group before filtering/pagination so a matching member never exposes siblings as cards.
const workCatalog = `WITH visible AS (
 SELECT s.id,s.library_id,` + seriesTitle + ` AS title,s.title AS source_title,s.aliases,
 COALESCE(wm.work_id,0) AS work_id,COALESCE(w.title,` + seriesTitle + `) AS card_title,
 CASE WHEN wm.work_id IS NULL THEN 's'||s.id ELSE 'w'||wm.work_id END AS card_key,
 COALESCE(wm.position,0) AS position,
 (SELECT count(*) FROM tv_episodes e JOIN media_items m ON m.id=e.media_id WHERE e.series_id=s.id AND m.missing=0) AS episodes,
 (SELECT count(DISTINCT e.season_number) FROM tv_episodes e JOIN media_items m ON m.id=e.media_id WHERE e.series_id=s.id AND m.missing=0) AS seasons,
 (SELECT count(*) FROM tv_episodes e JOIN media_items m ON m.id=e.media_id JOIN playback_progress p ON p.media_id=m.id WHERE e.series_id=s.id AND m.missing=0 AND p.user_id=? AND p.watched=1) AS watched,
 (SELECT max(p.updated_at) FROM tv_episodes e JOIN media_items m ON m.id=e.media_id JOIN playback_progress p ON p.media_id=m.id WHERE e.series_id=s.id AND m.missing=0 AND p.user_id=? AND (p.position>0 OR p.watched=1)) AS recent
 FROM tv_series s LEFT JOIN work_members wm ON wm.series_id=s.id LEFT JOIN works w ON w.id=wm.work_id
 WHERE ` + seriesVisible + ` AND (?=0 OR s.library_id=?)
), cards AS (
 SELECT card_key,work_id,card_title,sum(episodes) episodes,sum(seasons) seasons,sum(watched) watched,count(*) members,max(recent) recent,
 (SELECT id FROM visible v2 WHERE v2.card_key=v.card_key ORDER BY position,id LIMIT 1) id
 FROM visible v
 WHERE (?='' OR card_title LIKE ? ESCAPE '\' OR EXISTS(
 SELECT 1 FROM work_members search_wm JOIN tv_series s ON s.id=search_wm.series_id
 WHERE search_wm.work_id=v.work_id AND (` + seriesTitle + ` LIKE ? ESCAPE '\' OR s.title LIKE ? ESCAPE '\' OR s.aliases LIKE ? ESCAPE '\'))
 OR (work_id=0 AND (title LIKE ? ESCAPE '\' OR source_title LIKE ? ESCAPE '\' OR aliases LIKE ? ESCAPE '\')))
 GROUP BY card_key
) `

func (s *Store) SeriesPage(ctx context.Context, filter media.Filter, userID int64) (television.Page, error) {
	page := television.Page{Items: []television.Series{}, Page: filter.Page, PageSize: filter.PageSize}
	query := "%" + strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`).Replace(filter.Query) + "%"
	args := []any{userID, userID, filter.LibraryID, filter.LibraryID, filter.Query, query, query, query, query, query, query, query}
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return page, err
	}
	defer tx.Rollback()
	if err := tx.QueryRowContext(ctx, workCatalog+`SELECT count(*) FROM cards`, args...).Scan(&page.Total); err != nil {
		return page, err
	}
	args = append(args, filter.PageSize, (filter.Page-1)*filter.PageSize)
	rows, err := tx.QueryContext(ctx, workCatalog+`SELECT id,work_id,card_title,episodes,seasons,watched,members FROM cards ORDER BY recent DESC,card_title COLLATE NOCASE,id LIMIT ? OFFSET ?`, args...)
	if err != nil {
		return page, err
	}
	for rows.Next() {
		item := television.Series{}
		if err := rows.Scan(&item.ID, &item.WorkID, &item.Title, &item.EpisodeCount, &item.SeasonCount, &item.WatchedCount, &item.MemberCount); err != nil {
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
	for i, aggregate := range page.Items {
		item, err := scanSeries(tx.QueryRowContext(ctx, `SELECT `+seriesColumns+seriesFrom+` WHERE s.id=?`, aggregate.ID))
		if err != nil {
			return page, err
		}
		if aggregate.WorkID != 0 {
			item.Title = aggregate.Title
			item.Metadata = nil
			item.MemberCount = aggregate.MemberCount
		}
		item.EpisodeCount, item.SeasonCount, item.WatchedCount = aggregate.EpisodeCount, aggregate.SeasonCount, aggregate.WatchedCount
		page.Items[i] = item
	}
	return page, tx.Commit()
}
