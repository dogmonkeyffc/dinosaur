package store

import (
	"context"
	"encoding/json"
	"path/filepath"
	"strings"

	"dinosaur/internal/metadata"
)

func (s *Store) MetadataRecords(ctx context.Context, filter metadata.ReportFilter) (metadata.RecordPage, error) {
	page := metadata.RecordPage{Items: []metadata.Record{}, Summary: map[string]int{}, Page: filter.Page, PageSize: filter.PageSize}
	where := ` WHERE 1=1`
	args := []any{}
	if filter.LibraryID > 0 {
		where += ` AND r.library_id=?`
		args = append(args, filter.LibraryID)
	}
	if filter.Query != "" {
		where += ` AND (r.title LIKE ? ESCAPE '\' OR r.local_title LIKE ? ESCAPE '\'
   OR r.original_title LIKE ? ESCAPE '\' OR r.aliases LIKE ? ESCAPE '\' OR r.filename LIKE ? ESCAPE '\'
   OR (r.media_type='tv' AND EXISTS(SELECT 1 FROM tv_episodes e JOIN media_items m ON m.id=e.media_id
    WHERE e.series_id=r.series_id AND m.missing=0 AND m.relative_path LIKE ? ESCAPE '\')))`
		query := "%" + strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`).Replace(filter.Query) + "%"
		for range 6 {
			args = append(args, query)
		}
	}
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return page, err
	}
	defer tx.Rollback()
	rows, err := tx.QueryContext(ctx, metadataRecordSources+`SELECT `+recordStatus+`,count(*),sum(CASE WHEN (tmdb_id>0 OR bangumi_id>0) AND message<>'' THEN 1 ELSE 0 END)`+` FROM sources r`+where+` GROUP BY 1`, args...)
	if err != nil {
		return page, err
	}
	for rows.Next() {
		var status string
		var count, warnings int
		if err := rows.Scan(&status, &count, &warnings); err != nil {
			rows.Close()
			return page, err
		}
		page.Summary[status] = count
		page.Summary["all"] += count
		page.Summary["warning"] += warnings
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return page, err
	}
	rows.Close()
	page.Total = page.Summary["all"]
	if filter.Status == "warning" {
		where += ` AND (tmdb_id>0 OR bangumi_id>0) AND message<>''`
		page.Total = page.Summary["warning"]
	} else if filter.Status != "" {
		where += ` AND (` + recordStatus + `)=?`
		args = append(args, filter.Status)
		page.Total = page.Summary[filter.Status]
	}
	args = append(args, filter.PageSize, (filter.Page-1)*filter.PageSize)
	rows, err = tx.QueryContext(ctx, metadataRecordSources+`SELECT r.media_id,r.series_id,r.library_id,r.library_name,r.title,r.filename,
 r.tmdb_id,r.language,`+recordStatus+`,r.message,r.updated_at,r.candidates_json,r.bangumi_id
 FROM sources r`+where+` ORDER BY r.media_type,r.sort_id LIMIT ? OFFSET ?`, args...)
	if err != nil {
		return page, err
	}
	for rows.Next() {
		var item metadata.Record
		var candidates string
		if err := rows.Scan(&item.MediaID, &item.SeriesID, &item.LibraryID, &item.LibraryName, &item.Title, &item.Filename, &item.TMDBID, &item.Language, &item.Status, &item.Message, &item.UpdatedAt, &candidates, &item.BangumiID); err != nil {
			rows.Close()
			return page, err
		}
		if err := json.Unmarshal([]byte(candidates), &item.Candidates); err != nil {
			rows.Close()
			return page, err
		}
		item.Filename = filepath.Base(item.Filename)
		page.Items = append(page.Items, item)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return page, err
	}
	rows.Close()
	return page, tx.Commit()
}
