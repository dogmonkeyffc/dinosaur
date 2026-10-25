package store

import (
	"context"
	"time"

	"dinosaur/internal/media"
	"dinosaur/internal/metadata"
)

func (s *Store) SaveMovieMetadata(ctx context.Context, id int64, movie metadata.Movie) error {
	_, err := s.DB.ExecContext(ctx, `INSERT INTO movie_metadata
 (media_id,tmdb_id,title,original_title,overview,year,release_date,language,poster_key,aliases,status,message,updated_at)
 VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?) ON CONFLICT(media_id) DO UPDATE SET
 tmdb_id=excluded.tmdb_id,title=excluded.title,original_title=excluded.original_title,overview=excluded.overview,
 year=excluded.year,release_date=excluded.release_date,language=excluded.language,poster_key=excluded.poster_key,
 aliases=excluded.aliases,status=excluded.status,message=excluded.message,updated_at=excluded.updated_at`,
		id, movie.TMDBID, movie.Title, movie.OriginalTitle, movie.Overview, movie.Year, movie.ReleaseDate,
		movie.Language, movie.PosterKey, movie.Aliases, movie.Status, movie.Message, time.Now().Unix())
	return err
}

// RecordScrapeFailure preserves previously identified information on refresh failure.
func (s *Store) RecordScrapeFailure(ctx context.Context, id int64, status, message string) error {
	_, err := s.DB.ExecContext(ctx, `INSERT INTO movie_metadata(media_id,status,message,updated_at)
 VALUES (?,?,?,?) ON CONFLICT(media_id) DO UPDATE SET
 status=CASE WHEN movie_metadata.tmdb_id>0 THEN movie_metadata.status ELSE excluded.status END,
 message=excluded.message,updated_at=excluded.updated_at`, id, status, message, time.Now().Unix())
	return err
}

func (s *Store) ScrapeTargets(ctx context.Context, request metadata.ScrapeRequest) ([]media.Item, error) {
	query := `SELECT ` + mediaColumns + mediaFrom + ` WHERE m.missing=0 AND l.type='movie'`
	args := []any{}
	if request.LibraryID > 0 {
		query += ` AND m.library_id=?`
		args = append(args, request.LibraryID)
	}
	if request.MediaID > 0 {
		query += ` AND m.id=?`
		args = append(args, request.MediaID)
	}
	if !request.Refresh {
		query += ` AND COALESCE(md.tmdb_id,0)=0`
	}
	rows, err := s.DB.QueryContext(ctx, query+` ORDER BY m.id`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]media.Item, 0)
	for rows.Next() {
		item, err := scanMedia(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}
