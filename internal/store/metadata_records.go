package store

import (
	"context"
	"encoding/json"
	"path/filepath"
	"strings"

	"dinosaur/internal/metadata"
)

const recordStatus = `CASE WHEN md.tmdb_id>0 THEN 'matched' WHEN md.status='failed' THEN 'failed' WHEN md.status='needs_review' THEN 'needs_review' ELSE 'unmatched' END`

func (s *Store) MetadataRecords(ctx context.Context, filter metadata.ReportFilter) (metadata.RecordPage, error) {
	page := metadata.RecordPage{Items: []metadata.Record{}, Summary: map[string]int{}, Page: filter.Page, PageSize: filter.PageSize}
	where := ` WHERE m.missing=0 AND l.type='movie'`
	args := []any{}
	if filter.LibraryID > 0 {
		where += ` AND m.library_id=?`
		args = append(args, filter.LibraryID)
	}
	if filter.Query != "" {
		where += ` AND (` + displayTitle + ` LIKE ? ESCAPE '\' OR m.relative_path LIKE ? ESCAPE '\' OR md.original_title LIKE ? ESCAPE '\' OR md.aliases LIKE ? ESCAPE '\')`
		query := "%" + strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`).Replace(filter.Query) + "%"
		for range 4 {
			args = append(args, query)
		}
	}
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return page, err
	}
	defer tx.Rollback()
	rows, err := tx.QueryContext(ctx, `SELECT `+recordStatus+`,count(*),sum(CASE WHEN md.tmdb_id>0 AND md.message<>'' THEN 1 ELSE 0 END)`+mediaFrom+where+` GROUP BY 1`, args...)
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
		where += ` AND md.tmdb_id>0 AND md.message<>''`
		page.Total = page.Summary["warning"]
	} else if filter.Status != "" {
		where += ` AND (` + recordStatus + `)=?`
		args = append(args, filter.Status)
		page.Total = page.Summary[filter.Status]
	}
	args = append(args, filter.PageSize, (filter.Page-1)*filter.PageSize)
	rows, err = tx.QueryContext(ctx, `SELECT m.id,m.library_id,l.name,`+displayTitle+`,m.relative_path,COALESCE(md.tmdb_id,0),COALESCE(md.language,''),`+recordStatus+`,COALESCE(md.message,''),COALESCE(md.updated_at,0),COALESCE(md.candidates_json,'[]')`+mediaFrom+where+` ORDER BY m.id LIMIT ? OFFSET ?`, args...)
	if err != nil {
		return page, err
	}
	for rows.Next() {
		var item metadata.Record
		var candidates string
		if err := rows.Scan(&item.MediaID, &item.LibraryID, &item.LibraryName, &item.Title, &item.Filename, &item.TMDBID, &item.Language, &item.Status, &item.Message, &item.UpdatedAt, &candidates); err != nil {
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
