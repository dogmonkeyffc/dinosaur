package store

import (
	"context"
	"encoding/json"
	"path/filepath"
	"strings"

	"dinosaur/internal/media"
	"dinosaur/internal/metadata"
)

const mediaFrom = ` FROM media_items m JOIN libraries l ON l.id=m.library_id LEFT JOIN movie_metadata md ON md.media_id=m.id`
const displayTitle = `COALESCE(NULLIF(md.title,''),m.title)`
const mediaColumns = `m.id,m.library_id,l.name,m.relative_path,` + displayTitle + `,COALESCE(NULLIF(md.year,0),m.year),m.size,m.modified_ns,m.duration,m.format,m.video_codec,m.width,m.height,m.tracks_json,COALESCE(NULLIF(md.poster_key,''),m.poster_key),m.probe_error,m.missing,
 COALESCE(md.tmdb_id,0),COALESCE(md.original_title,''),COALESCE(md.overview,''),COALESCE(md.release_date,''),COALESCE(md.language,''),COALESCE(md.status,''),COALESCE(md.message,''),COALESCE(md.updated_at,0)`

func scanMedia(row interface{ Scan(...any) error }) (media.Item, error) {
	var item media.Item
	var tracks string
	var info metadata.Movie
	err := row.Scan(&item.ID, &item.LibraryID, &item.LibraryName, &item.RelativePath, &item.Title, &item.Year, &item.Size, &item.ModifiedNS, &item.Duration, &item.Format, &item.VideoCodec, &item.Width, &item.Height, &tracks, &item.PosterKey, &item.ProbeError, &item.Missing, &info.TMDBID, &info.OriginalTitle, &info.Overview, &info.ReleaseDate, &info.Language, &info.Status, &info.Message, &info.UpdatedAt)
	if err != nil {
		return item, err
	}
	item.Filename = filepath.Base(item.RelativePath)
	if info.Status != "" {
		info.Title, info.Year = item.Title, item.Year
		item.Metadata = &info
	}
	err = json.Unmarshal([]byte(tracks), &item.Tracks)
	return item, err
}
func (s *Store) MediaByID(ctx context.Context, id int64) (media.Item, error) {
	return scanMedia(s.DB.QueryRowContext(ctx, `SELECT `+mediaColumns+mediaFrom+` WHERE m.id=? AND m.missing=0`, id))
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
		conditions = append(conditions, `(m.title LIKE ? ESCAPE '\' OR md.title LIKE ? ESCAPE '\' OR md.original_title LIKE ? ESCAPE '\' OR md.aliases LIKE ? ESCAPE '\')`)
		escaped := strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`).Replace(filter.Query)
		for range 4 {
			args = append(args, "%"+escaped+"%")
		}
	}
	where := strings.Join(conditions, " AND ")
	// A read transaction keeps totals and page contents consistent during scanning.
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return result, err
	}
	defer tx.Rollback()
	if err := tx.QueryRowContext(ctx, `SELECT count(*)`+mediaFrom+` WHERE `+where, args...).Scan(&result.Total); err != nil {
		return result, err
	}
	queryArgs := append(append([]any{}, args...), filter.PageSize, (filter.Page-1)*filter.PageSize)
	rows, err := tx.QueryContext(ctx, `SELECT `+mediaColumns+mediaFrom+` WHERE `+where+` ORDER BY `+displayTitle+` COLLATE NOCASE,m.id LIMIT ? OFFSET ?`, queryArgs...)
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
