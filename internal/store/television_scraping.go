package store

import (
	"context"
	"fmt"

	"dinosaur/internal/media"
	"dinosaur/internal/metadata"
	"dinosaur/internal/television"
)

// TelevisionRepository presents series as metadata targets to the shared worker.
// These target IDs are series IDs, never playback IDs.
type TelevisionRepository struct{ *Store }

func seriesTarget(item television.Series) media.Item {
	filename := item.Title + ".mp4"
	if item.Year > 0 {
		filename = fmt.Sprintf("%s (%d).mp4", item.Title, item.Year)
	}
	return media.Item{ID: item.ID, LibraryID: item.LibraryID, LibraryName: item.LibraryName, Title: item.Title, Year: item.Year,
		Filename: filename, PosterKey: item.PosterKey, Metadata: item.Metadata, Type: "tv"}
}

func (s TelevisionRepository) MediaByID(ctx context.Context, id int64) (media.Item, error) {
	item, err := s.SeriesByID(ctx, id)
	return seriesTarget(item), err
}

func (s TelevisionRepository) ScrapeTargets(ctx context.Context, request metadata.ScrapeRequest) ([]media.Item, error) {
	query := `SELECT ` + seriesColumns + seriesFrom + ` WHERE ` + seriesVisible
	args := []any{}
	if request.LibraryID > 0 {
		query += ` AND s.library_id=?`
		args = append(args, request.LibraryID)
	}
	if request.MediaID > 0 {
		query += ` AND s.id=?`
		args = append(args, request.MediaID)
	}
	if !request.Refresh {
		query += ` AND COALESCE(json_extract(s.metadata_json,'$.tmdb_id'),0)=0 AND COALESCE(json_extract(s.metadata_json,'$.bangumi_id'),0)=0 AND COALESCE(json_extract(s.metadata_json,'$.source'),'') NOT IN ('agent','nfo')`
	}
	rows, err := s.DB.QueryContext(ctx, query+` ORDER BY s.id`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := []media.Item{}
	for rows.Next() {
		item, err := scanSeries(rows)
		if err != nil {
			return nil, err
		}
		// Matching always uses the local show title, even after a provider refresh.
		result = append(result, seriesTarget(item))
	}
	return result, rows.Err()
}

func (s TelevisionRepository) NewScrapeJob(ctx context.Context, request metadata.ScrapeRequest) (metadata.Job, error) {
	request.MediaType = "tv"
	return s.Store.NewScrapeJob(ctx, request)
}
