package store

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"dinosaur/internal/television"
)

type seriesCollection struct {
	libraryID int64
	key       string
}

type collectionEpisode struct {
	mediaID  int64
	seriesID int64
	parsed   television.EpisodeName
}

func migrateSeriesCollections(ctx context.Context, tx *sql.Tx) error {
	const version = "014_series_collections"
	var applied int
	if err := tx.QueryRowContext(ctx, `SELECT count(*) FROM schema_migrations WHERE version=?`, version).Scan(&applied); err != nil || applied != 0 {
		return err
	}
	rows, err := tx.QueryContext(ctx, `SELECT m.id,m.library_id,m.relative_path,e.series_id,s.source_key
 FROM tv_episodes e JOIN media_items m ON m.id=e.media_id JOIN tv_series s ON s.id=e.series_id
 WHERE s.source_key LIKE 'folder:%' ORDER BY e.season_number=0,e.season_number,e.episode_number,m.id`)
	if err != nil {
		return err
	}
	groups := make(map[seriesCollection][]collectionEpisode)
	for rows.Next() {
		var item collectionEpisode
		var libraryID int64
		var relative, key string
		if err := rows.Scan(&item.mediaID, &libraryID, &relative, &item.seriesID, &key); err != nil {
			rows.Close()
			return err
		}
		parsed, ok := television.ParseEpisode(relative)
		if ok && parsed.Collection && parsed.Key != key {
			item.parsed = parsed
			group := seriesCollection{libraryID, parsed.Key}
			groups[group] = append(groups[group], item)
		}
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return err
	}
	if err := rows.Close(); err != nil {
		return err
	}
	for group, episodes := range groups {
		if err := mergeSeriesCollection(ctx, tx, group, episodes); err != nil {
			return fmt.Errorf("group series %q: %w", group.key, err)
		}
	}
	_, err = tx.ExecContext(ctx, `INSERT INTO schema_migrations VALUES (?,?)`, version, time.Now().Unix())
	return err
}

func mergeSeriesCollection(ctx context.Context, tx *sql.Tx, group seriesCollection, episodes []collectionEpisode) error {
	var seriesID int64
	err := tx.QueryRowContext(ctx, `SELECT id FROM tv_series WHERE library_id=? AND source_key=?`, group.libraryID, group.key).Scan(&seriesID)
	if err == sql.ErrNoRows {
		first := episodes[0]
		seriesID = first.seriesID
		// Use the collection title while retaining the first season's existing artwork and other metadata.
		// Other season rows retain their metadata and history but disappear from browsing once empty.
		_, err = tx.ExecContext(ctx, `UPDATE tv_series SET source_key=?,title=?,year=CASE WHEN ?>0 THEN ? ELSE year END,
 metadata_json=json_remove(metadata_json,'$.title'),metadata_version=metadata_version+1 WHERE id=?`,
			group.key, first.parsed.Series, first.parsed.Year, first.parsed.Year, seriesID)
	}
	if err != nil {
		return err
	}
	for _, episode := range episodes {
		if episode.seriesID == seriesID {
			continue
		}
		if _, err := tx.ExecContext(ctx, `UPDATE tv_episodes SET series_id=? WHERE media_id=?`, seriesID, episode.mediaID); err != nil {
			return err
		}
	}
	return nil
}
