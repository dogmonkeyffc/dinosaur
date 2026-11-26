package store

import (
	"context"
	"time"

	"dinosaur/internal/metadata"
)

const seriesIMDb = `COALESCE(json_extract(s.metadata_json,'$.ids.imdb'),'')`
const directSeriesTMDB = `COALESCE(NULLIF(CAST(json_extract(s.metadata_json,'$.tmdb_id') AS INTEGER),0),
	NULLIF(CAST(json_extract(s.metadata_json,'$.ids.tmdb') AS INTEGER),0),0)`
const seriesTMDB = `COALESCE(NULLIF(` + directSeriesTMDB + `,0),
	(SELECT x.tmdb_id FROM season_poster_matches x WHERE x.imdb_id=` + seriesIMDb + `),0)`

func (s *Store) PendingSeasonPosterMatches(ctx context.Context, limit int) ([]string, error) {
	rows, err := s.DB.QueryContext(ctx, `SELECT DISTINCT `+seriesIMDb+`
		FROM tv_series s LEFT JOIN season_poster_matches x ON x.imdb_id=`+seriesIMDb+`
		WHERE `+directSeriesTMDB+`<=0 AND `+seriesIMDb+`<>'' AND `+seriesVisible+`
		AND COALESCE(x.tmdb_id,0)=0 AND COALESCE(x.retry_at,0)<=? ORDER BY 1 LIMIT ?`, time.Now().Unix(), limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

func (s *Store) SaveSeasonPosterMatch(ctx context.Context, imdb string, id int64, retry time.Time) error {
	_, err := s.DB.ExecContext(ctx, `INSERT INTO season_poster_matches(imdb_id,tmdb_id,retry_at) VALUES(?,?,?)
		ON CONFLICT(imdb_id) DO UPDATE SET tmdb_id=excluded.tmdb_id,retry_at=excluded.retry_at
		WHERE season_poster_matches.tmdb_id=0`, imdb, id, retry.Unix())
	return err
}

func (s *Store) PendingSeasonPosters(ctx context.Context, limit int) ([]metadata.SeasonPosterTarget, error) {
	rows, err := s.DB.QueryContext(ctx, `WITH targets AS (
		SELECT DISTINCT `+seriesTMDB+` AS tmdb_id,e.season_number
		FROM tv_series s JOIN tv_episodes e ON e.series_id=s.id
		JOIN media_items m ON m.id=e.media_id
		LEFT JOIN season_posters p ON p.tmdb_id=`+seriesTMDB+` AND p.season_number=e.season_number
		WHERE `+seriesTMDB+`>0 AND m.missing=0 AND e.season_number>=0
		AND COALESCE(p.poster_key,'')='' AND COALESCE(p.retry_at,0)<=?
	), batch AS (SELECT DISTINCT tmdb_id FROM targets ORDER BY tmdb_id LIMIT ?)
	SELECT t.tmdb_id,t.season_number FROM targets t JOIN batch b ON b.tmdb_id=t.tmdb_id
	ORDER BY t.tmdb_id,t.season_number`, time.Now().Unix(), limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var targets []metadata.SeasonPosterTarget
	for rows.Next() {
		var id int64
		var number int
		if err := rows.Scan(&id, &number); err != nil {
			return nil, err
		}
		if len(targets) == 0 || targets[len(targets)-1].TMDBID != id {
			targets = append(targets, metadata.SeasonPosterTarget{TMDBID: id})
		}
		target := &targets[len(targets)-1]
		target.Numbers = append(target.Numbers, number)
	}
	return targets, rows.Err()
}

func (s *Store) SaveSeasonPoster(ctx context.Context, id int64, number int, key string, retry time.Time) error {
	_, err := s.DB.ExecContext(ctx, `INSERT INTO season_posters(tmdb_id,season_number,poster_key,retry_at)
		VALUES(?,?,?,?) ON CONFLICT(tmdb_id,season_number) DO UPDATE SET
		poster_key=excluded.poster_key,retry_at=excluded.retry_at WHERE season_posters.poster_key=''`, id, number, key, retry.Unix())
	return err
}

func (s *Store) SeriesSeasonCovers(ctx context.Context, id int64) (map[int]string, error) {
	rows, err := s.DB.QueryContext(ctx, `SELECT DISTINCT e.season_number,COALESCE(NULLIF(p.poster_key,''),s.poster_key)
		FROM tv_series s JOIN tv_episodes e ON e.series_id=s.id JOIN media_items m ON m.id=e.media_id
		LEFT JOIN season_posters p ON p.tmdb_id=`+seriesTMDB+` AND p.season_number=e.season_number
		WHERE s.id=? AND m.missing=0 AND `+libraryVisible("s.library_id"), id, libraryViewer(ctx))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	covers := map[int]string{}
	for rows.Next() {
		var number int
		var key string
		if err := rows.Scan(&number, &key); err != nil {
			return nil, err
		}
		covers[number] = key
	}
	return covers, rows.Err()
}
