package store

import (
	"context"
	"encoding/json"
	"time"

	"dinosaur/internal/metadata"
)

func (s TelevisionRepository) SaveMovieMetadata(ctx context.Context, id int64, info metadata.Movie) error {
	info.UpdatedAt = time.Now().Unix()
	raw, err := json.Marshal(info)
	if err != nil {
		return err
	}
	_, err = s.DB.ExecContext(ctx, `UPDATE tv_series SET metadata_json=?,poster_key=?,aliases=? WHERE id=?`, string(raw), info.PosterKey, info.Aliases, id)
	return err
}

func (s TelevisionRepository) RecordScrapeFailure(ctx context.Context, id int64, status, message string) error {
	_, err := s.DB.ExecContext(ctx, `UPDATE tv_series SET metadata_json=json_set(metadata_json,
 '$.status',CASE WHEN (COALESCE(json_extract(metadata_json,'$.tmdb_id'),0)>0 OR COALESCE(json_extract(metadata_json,'$.bangumi_id'),0)>0) THEN json_extract(metadata_json,'$.status') ELSE ? END,
 '$.message',?,'$.updated_at',?) WHERE id=?`, status, message, time.Now().Unix(), id)
	return err
}

func (s TelevisionRepository) RecordScrapeReview(ctx context.Context, id int64, candidates []metadata.Candidate) error {
	message := "ambiguous"
	if len(candidates) == 0 {
		message = "not_found"
	}
	raw, err := json.Marshal(candidates)
	if err != nil {
		return err
	}
	_, err = s.DB.ExecContext(ctx, `UPDATE tv_series SET metadata_json=json_set(metadata_json,'$.status','needs_review',
 '$.message',?,'$.candidates',json(?),'$.updated_at',?) WHERE id=? AND COALESCE(json_extract(metadata_json,'$.tmdb_id'),0)=0 AND COALESCE(json_extract(metadata_json,'$.bangumi_id'),0)=0`, message, string(raw), time.Now().Unix(), id)
	return err
}
