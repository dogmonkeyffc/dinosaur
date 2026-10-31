package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"dinosaur/internal/agent"
	"dinosaur/internal/metadata"
)

func (s *Store) AgentTarget(ctx context.Context, kind string, id int64) (agent.Target, error) {
	target := agent.Target{Kind: kind, ID: id, Files: []agent.File{}}
	if kind == "movie" {
		item, err := s.MediaByID(ctx, id)
		if err != nil {
			return target, err
		}
		if item.Type != "movie" {
			return target, sql.ErrNoRows
		}
		target.LibraryID, target.LibraryName, target.Title, target.Year, target.Metadata = item.LibraryID, item.LibraryName, item.Title, item.Year, item.Metadata
		target.FileCount = 1
		target.Files = append(target.Files, agent.File{Path: item.RelativePath, Size: item.Size, Duration: item.Duration, VideoCodec: item.VideoCodec, Width: item.Width, Height: item.Height, Tracks: item.Tracks})
	} else if kind == "tv" {
		item, err := s.SeriesByID(ctx, id)
		if err != nil {
			return target, err
		}
		target.LibraryID, target.LibraryName, target.Title, target.Year, target.Metadata = item.LibraryID, item.LibraryName, item.Title, item.Year, item.Metadata
		target.FileCount = item.EpisodeCount
		rows, err := s.DB.QueryContext(ctx, `SELECT m.relative_path,m.size,m.duration,m.video_codec,m.width,m.height,m.tracks_json,e.season_number,e.episode_number
 FROM tv_episodes e JOIN media_items m ON m.id=e.media_id WHERE e.series_id=? AND m.missing=0 ORDER BY e.season_number,e.episode_number,m.id LIMIT 1000`, id)
		if err != nil {
			return target, err
		}
		for rows.Next() {
			var file agent.File
			var tracks string
			if err := rows.Scan(&file.Path, &file.Size, &file.Duration, &file.VideoCodec, &file.Width, &file.Height, &tracks, &file.Season, &file.Episode); err != nil {
				rows.Close()
				return target, err
			}
			if err := json.Unmarshal([]byte(tracks), &file.Tracks); err != nil {
				rows.Close()
				return target, err
			}
			target.Files = append(target.Files, file)
		}
		err = rows.Err()
		rows.Close()
		if err != nil {
			return target, err
		}
	} else {
		return target, agent.ErrInput
	}
	if kind == "movie" {
		err := s.DB.QueryRowContext(ctx, `SELECT COALESCE((SELECT revision FROM movie_metadata WHERE media_id=?),0)`, id).Scan(&target.Version)
		if err != nil {
			return target, err
		}
	} else {
		err := s.DB.QueryRowContext(ctx, `SELECT metadata_version FROM tv_series WHERE id=?`, id).Scan(&target.Version)
		if err != nil {
			return target, err
		}
	}
	if target.Metadata != nil {
		target.Metadata.Version = target.Version
	}
	library, err := s.LibraryByID(ctx, target.LibraryID)
	if err != nil {
		return target, err
	}
	target.LibraryRoot = library.RootPath
	return target, nil
}

func (s *Store) ApplyAgentMetadata(ctx context.Context, key agent.Key, kind string, id, expected int64, info metadata.Movie) (int64, error) {
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()
	var libraryID, version int64
	var poster string
	if kind == "movie" {
		err = tx.QueryRowContext(ctx, `SELECT m.library_id,COALESCE(md.revision,0),COALESCE(NULLIF(md.poster_key,''),m.poster_key)
 `+mediaFrom+` WHERE m.id=? AND m.missing=0 AND l.type='movie'`, id).Scan(&libraryID, &version, &poster)
	} else if kind == "tv" {
		err = tx.QueryRowContext(ctx, `SELECT s.library_id,s.metadata_version,s.poster_key `+seriesFrom+` WHERE s.id=? AND `+seriesVisible, id).Scan(&libraryID, &version, &poster)
	} else {
		return 0, agent.ErrInput
	}
	if err != nil {
		return 0, err
	}
	if !key.Allows(libraryID, true) {
		return 0, agent.ErrForbidden
	}
	if expected != version {
		return version, metadata.ErrConflict
	}
	if !info.Managed() {
		return 0, agent.ErrInput
	}
	if info.PosterKey == "" {
		info.PosterKey = poster
	}
	if kind == "movie" {
		err = saveMovieMetadata(ctx, tx, id, info)
	} else {
		err = saveSeriesMetadata(ctx, tx, id, info)
	}
	if err != nil {
		return 0, err
	}
	_, err = tx.ExecContext(ctx, `INSERT INTO agent_metadata_events(key_id,key_name,kind,target_id,revision,source,created_at) VALUES (?,?,?,?,?,?,?)`, key.ID, key.Name, kind, id, version+1, info.Source, time.Now().Unix())
	if err != nil {
		return 0, err
	}
	return version + 1, tx.Commit()
}
