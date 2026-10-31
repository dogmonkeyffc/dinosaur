package store

import (
	"context"
	"encoding/json"
	"time"

	"dinosaur/internal/media"
	"dinosaur/internal/metadata"
)

func (s *Store) SaveMovieMetadata(ctx context.Context, id int64, movie metadata.Movie) error {
	candidates, err := json.Marshal(movie.Candidates)
	if err != nil {
		return err
	}
	_, err = s.DB.ExecContext(ctx, `INSERT INTO movie_metadata
 (media_id,tmdb_id,title,original_title,overview,year,release_date,language,poster_key,aliases,status,message,updated_at,candidates_json,original_language,bangumi_id)
 VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?) ON CONFLICT(media_id) DO UPDATE SET
 tmdb_id=excluded.tmdb_id,title=excluded.title,original_title=excluded.original_title,overview=excluded.overview,
 year=excluded.year,release_date=excluded.release_date,language=excluded.language,poster_key=excluded.poster_key,
 aliases=excluded.aliases,status=excluded.status,message=excluded.message,updated_at=excluded.updated_at,candidates_json=excluded.candidates_json,original_language=excluded.original_language,bangumi_id=excluded.bangumi_id`,
		id, movie.TMDBID, movie.Title, movie.OriginalTitle, movie.Overview, movie.Year, movie.ReleaseDate,
		movie.Language, movie.PosterKey, movie.Aliases, movie.Status, movie.Message, time.Now().Unix(), string(candidates), movie.OriginalLanguage, movie.BangumiID)
	return err
}

// RecordScrapeFailure preserves previously identified information on refresh failure.
func (s *Store) RecordScrapeFailure(ctx context.Context, id int64, status, message string) error {
	_, err := s.DB.ExecContext(ctx, `INSERT INTO movie_metadata(media_id,status,message,updated_at)
 VALUES (?,?,?,?) ON CONFLICT(media_id) DO UPDATE SET
 status=CASE WHEN (movie_metadata.tmdb_id>0 OR movie_metadata.bangumi_id>0) THEN movie_metadata.status ELSE excluded.status END,
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
		query += ` AND COALESCE(md.tmdb_id,0)=0 AND COALESCE(md.bangumi_id,0)=0`
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

// RecordScrapeReview caches suggestions without overwriting a concurrent match.
func (s *Store) RecordScrapeReview(ctx context.Context, id int64, candidates []metadata.Candidate) error {
	message := "ambiguous"
	if len(candidates) == 0 {
		message = "not_found"
	}
	data, err := json.Marshal(candidates)
	if err != nil {
		return err
	}
	_, err = s.DB.ExecContext(ctx, `INSERT INTO movie_metadata(media_id,status,message,candidates_json,updated_at)
 VALUES (?,'needs_review',?,?,?) ON CONFLICT(media_id) DO UPDATE SET
 status=excluded.status,message=excluded.message,candidates_json=excluded.candidates_json,updated_at=excluded.updated_at
 WHERE movie_metadata.tmdb_id=0 AND movie_metadata.bangumi_id=0`, id, message, string(data), time.Now().Unix())
	return err
}
