package store

import (
	"context"
	"encoding/json"
	"path/filepath"
	"strings"

	"dinosaur/internal/media"
)

const mediaColumns = `m.id,m.library_id,l.name,m.relative_path,m.title,m.year,m.size,m.modified_ns,m.duration,m.format,m.video_codec,m.width,m.height,m.tracks_json,m.poster_key,m.probe_error,m.missing`

func scanMedia(row interface{ Scan(...any) error }) (media.Item, error) {
	var item media.Item
	var tracks string
	err := row.Scan(&item.ID, &item.LibraryID, &item.LibraryName, &item.RelativePath, &item.Title, &item.Year, &item.Size, &item.ModifiedNS, &item.Duration, &item.Format, &item.VideoCodec, &item.Width, &item.Height, &tracks, &item.PosterKey, &item.ProbeError, &item.Missing)
	if err != nil {
		return item, err
	}
	item.Filename = filepath.Base(item.RelativePath)
	err = json.Unmarshal([]byte(tracks), &item.Tracks)
	return item, err
}
func (s *Store) MediaByID(ctx context.Context, id int64) (media.Item, error) {
	return scanMedia(s.DB.QueryRowContext(ctx, `SELECT `+mediaColumns+` FROM media_items m JOIN libraries l ON l.id=m.library_id WHERE m.id=? AND m.missing=0`, id))
}
func (s *Store) MediaPage(ctx context.Context, filter media.Filter) (media.Page, error) {
	result := media.Page{Items: make([]media.Item, 0), Page: filter.Page, PageSize: filter.PageSize}
	conditions := []string{"m.missing=0", "l.type='movie'"}
	args := make([]any, 0)
	if filter.LibraryID > 0 {
		conditions = append(conditions, "m.library_id=?")
		args = append(args, filter.LibraryID)
	}
	if filter.Query != "" {
		conditions = append(conditions, `m.title LIKE ? ESCAPE '\'`)
		escaped := strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`).Replace(filter.Query)
		args = append(args, "%"+escaped+"%")
	}
	where := strings.Join(conditions, " AND ")
	// A read transaction keeps totals and page contents consistent during scanning.
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return result, err
	}
	defer tx.Rollback()
	if err := tx.QueryRowContext(ctx, `SELECT count(*) FROM media_items m JOIN libraries l ON l.id=m.library_id WHERE `+where, args...).Scan(&result.Total); err != nil {
		return result, err
	}
	queryArgs := append(append([]any{}, args...), filter.PageSize, (filter.Page-1)*filter.PageSize)
	rows, err := tx.QueryContext(ctx, `SELECT `+mediaColumns+` FROM media_items m JOIN libraries l ON l.id=m.library_id WHERE `+where+` ORDER BY m.title COLLATE NOCASE,m.id LIMIT ? OFFSET ?`, queryArgs...)
	if err != nil {
		return result, err
	}
	for rows.Next() {
		item, err := scanMedia(rows)
		if err != nil {
			rows.Close()
			return result, err
		}
		result.Items = append(result.Items, item)
	}
	rowsErr := rows.Err()
	if err := rows.Close(); err != nil {
		return result, err
	}
	if rowsErr != nil {
		return result, rowsErr
	}
	return result, tx.Commit()
}
