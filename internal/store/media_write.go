package store

import (
	"context"
	"database/sql"
	"encoding/json"

	"dinosaur/internal/media"
)

func (s *Store) MediaSnapshot(ctx context.Context, libraryID int64) (map[string]media.Item, error) {
	rows, err := s.DB.QueryContext(ctx, `SELECT id,relative_path,size,modified_ns,poster_key,probe_error,duration FROM media_items WHERE library_id=?`, libraryID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make(map[string]media.Item)
	for rows.Next() {
		var item media.Item
		item.LibraryID = libraryID
		if err := rows.Scan(&item.ID, &item.RelativePath, &item.Size, &item.ModifiedNS, &item.PosterKey, &item.ProbeError, &item.Duration); err != nil {
			return nil, err
		}
		result[item.RelativePath] = item
	}
	return result, rows.Err()
}
func (s *Store) SaveMedia(ctx context.Context, item media.Item, jobID int64) error {
	return saveMedia(ctx, s.DB, item, jobID)
}

type mediaWriter interface {
	ExecContext(context.Context, string, ...any) (sql.Result, error)
}

func saveMedia(ctx context.Context, writer mediaWriter, item media.Item, jobID int64) error {
	if item.Tracks == nil {
		item.Tracks = make([]media.Track, 0)
	}
	tracks, err := json.Marshal(item.Tracks)
	if err != nil {
		return err
	}
	_, err = writer.ExecContext(ctx, `INSERT INTO media_items
 (library_id,relative_path,title,year,size,modified_ns,duration,format,video_codec,width,height,tracks_json,poster_key,probe_error,last_seen_scan)
 VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?) ON CONFLICT(library_id,relative_path) DO UPDATE SET
 title=excluded.title,year=excluded.year,size=excluded.size,modified_ns=excluded.modified_ns,duration=excluded.duration,
 format=excluded.format,video_codec=excluded.video_codec,width=excluded.width,height=excluded.height,
 tracks_json=excluded.tracks_json,poster_key=excluded.poster_key,probe_error=excluded.probe_error,missing=0,last_seen_scan=excluded.last_seen_scan`,
		item.LibraryID, item.RelativePath, item.Title, item.Year, item.Size, item.ModifiedNS, item.Duration, item.Format, item.VideoCodec, item.Width, item.Height, string(tracks), item.PosterKey, item.ProbeError, jobID)
	return err
}
func (s *Store) TouchMedia(ctx context.Context, id, jobID int64) error {
	_, err := s.DB.ExecContext(ctx, `UPDATE media_items SET last_seen_scan=?,missing=0 WHERE id=?`, jobID, id)
	return err
}
func (s *Store) MarkMissingMedia(ctx context.Context, libraryID, jobID int64) error {
	_, err := s.DB.ExecContext(ctx, `UPDATE media_items SET missing=1 WHERE library_id=? AND last_seen_scan<>?`, libraryID, jobID)
	return err
}

func (s *Store) SetMediaPoster(ctx context.Context, id int64, key string) error {
	_, err := s.DB.ExecContext(ctx, `UPDATE media_items SET poster_key=? WHERE id=?`, key, id)
	return err
}
