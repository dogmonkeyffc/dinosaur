package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"dinosaur/internal/metadata"
)

type metadataExecutor interface {
	ExecContext(context.Context, string, ...any) (sql.Result, error)
}

func saveMovieMetadata(ctx context.Context, db metadataExecutor, id int64, movie metadata.Movie) error {
	candidates, err := json.Marshal(movie.Candidates)
	if err != nil {
		return err
	}
	ids, err := json.Marshal(movie.ExternalIDs)
	if err != nil {
		return err
	}
	result, err := db.ExecContext(ctx, `INSERT INTO movie_metadata
 (media_id,tmdb_id,title,original_title,overview,year,release_date,language,poster_key,aliases,status,message,updated_at,candidates_json,original_language,bangumi_id,source,source_url,external_ids_json,revision)
 VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,1) ON CONFLICT(media_id) DO UPDATE SET
 tmdb_id=excluded.tmdb_id,title=excluded.title,original_title=excluded.original_title,overview=excluded.overview,
 year=excluded.year,release_date=excluded.release_date,language=excluded.language,poster_key=excluded.poster_key,
 aliases=excluded.aliases,status=excluded.status,message=excluded.message,updated_at=excluded.updated_at,
 candidates_json=excluded.candidates_json,original_language=excluded.original_language,bangumi_id=excluded.bangumi_id,
 source=excluded.source,source_url=excluded.source_url,external_ids_json=excluded.external_ids_json,revision=movie_metadata.revision+1
 WHERE movie_metadata.source NOT IN ('agent','nfo') OR excluded.source IN ('agent','nfo')`,
		id, movie.TMDBID, movie.Title, movie.OriginalTitle, movie.Overview, movie.Year, movie.ReleaseDate, movie.Language, movie.PosterKey, movie.Aliases, movie.Status, movie.Message, time.Now().Unix(), string(candidates), movie.OriginalLanguage, movie.BangumiID, movie.Source, movie.SourceURL, string(ids))
	return savedMetadata(result, err)
}

func saveSeriesMetadata(ctx context.Context, db metadataExecutor, id int64, info metadata.Movie) error {
	info.UpdatedAt = time.Now().Unix()
	raw, err := json.Marshal(info)
	if err != nil {
		return err
	}
	result, err := db.ExecContext(ctx, `UPDATE tv_series SET metadata_json=json_set(?,'$.version',metadata_version+1),poster_key=?,aliases=?,metadata_version=metadata_version+1 WHERE id=?
 AND (COALESCE(json_extract(metadata_json,'$.source'),'') NOT IN ('agent','nfo') OR ? IN ('agent','nfo'))`, string(raw), info.PosterKey, info.Aliases, id, info.Source)
	return savedMetadata(result, err)
}

func savedMetadata(result sql.Result, err error) error {
	if err != nil {
		return err
	}
	count, err := result.RowsAffected()
	if err == nil && count == 0 {
		return metadata.ErrConflict
	}
	return err
}
