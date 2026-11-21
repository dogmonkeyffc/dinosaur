package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"dinosaur/internal/playmetrics"
)

const (
	playbackMetricsRetention = 30 * 24 * time.Hour
	playbackMetricsMaxRows   = 10_000
	playbackMetricsListLimit = 100
)

func (s *Store) SavePlaybackMetrics(ctx context.Context, userID int64, snapshot playmetrics.Snapshot) error {
	if err := snapshot.Validate(); err != nil {
		return err
	}
	var streamID *string
	if snapshot.StreamID != nil {
		id := strings.ToLower(*snapshot.StreamID)
		streamID = &id
	}
	var sourceJSON *string
	if snapshot.Source != nil {
		data, err := json.Marshal(snapshot.Source)
		if err != nil {
			return fmt.Errorf("encode playback metrics source: %w", err)
		}
		encoded := string(data)
		sourceJSON = &encoded
	}
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	now := time.Now()
	// Expire first so expired IDs cannot recover old cumulative data.
	if err := expirePlaybackMetrics(ctx, tx, now); err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, savePlaybackMetricsSQL, userID, strings.ToLower(snapshot.ID), streamID,
		snapshot.Mode, snapshot.Quality, snapshot.BurnSubtitle, snapshot.Width, snapshot.Height, snapshot.StartupMS, snapshot.Stalls,
		snapshot.StallMS, snapshot.PlayedMS, snapshot.BufferMS, snapshot.DroppedFrames, snapshot.TotalFrames, now.UnixNano(),
		sourceJSON, snapshot.Reason)
	if err != nil {
		return err
	}
	if err := capPlaybackMetrics(ctx, tx); err != nil {
		return err
	}
	return tx.Commit()
}

// Keep counters monotonic and preserve optional values if subsequent reports omit
// them. Source and reason are independently set once, when first known.
// A report with any regressing core counter cannot replace current gauges.
// With no sequence in the protocol, equal counters have arrival-order semantics.
const savePlaybackMetricsSQL = `
INSERT INTO playback_metrics(user_id,id,stream_id,mode,quality,burn_subtitle,width,height,startup_ms,stalls,stall_ms,played_ms,buffer_ms,dropped_frames,total_frames,updated_at_ns,source_json,reason)
VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)
ON CONFLICT(user_id,id) DO UPDATE SET
  source_json=coalesce(playback_metrics.source_json, excluded.source_json),
  reason=coalesce(nullif(playback_metrics.reason, ''), excluded.reason),
  startup_ms=coalesce(playback_metrics.startup_ms, excluded.startup_ms),
  stalls=max(playback_metrics.stalls, excluded.stalls),
  stall_ms=max(playback_metrics.stall_ms, excluded.stall_ms),
  played_ms=max(playback_metrics.played_ms, excluded.played_ms),
  dropped_frames=` + mergeDroppedFramesSQL + `,
  total_frames=` + mergeTotalFramesSQL + `,
  stream_id=CASE WHEN ` + currentPlaybackReportSQL + ` THEN coalesce(excluded.stream_id, playback_metrics.stream_id) ELSE playback_metrics.stream_id END,
  mode=CASE WHEN ` + currentPlaybackReportSQL + ` THEN excluded.mode ELSE playback_metrics.mode END,
  quality=CASE WHEN ` + currentPlaybackReportSQL + ` THEN excluded.quality ELSE playback_metrics.quality END,
  burn_subtitle=CASE WHEN ` + currentPlaybackReportSQL + ` THEN coalesce(excluded.burn_subtitle, playback_metrics.burn_subtitle)
    ELSE coalesce(playback_metrics.burn_subtitle, excluded.burn_subtitle) END,
  width=CASE WHEN ` + currentPlaybackReportSQL + ` THEN coalesce(nullif(excluded.width, 0), playback_metrics.width) ELSE playback_metrics.width END,
  height=CASE WHEN ` + currentPlaybackReportSQL + ` THEN coalesce(nullif(excluded.height, 0), playback_metrics.height) ELSE playback_metrics.height END,
  buffer_ms=CASE WHEN ` + currentPlaybackReportSQL + ` THEN excluded.buffer_ms ELSE playback_metrics.buffer_ms END,
  updated_at_ns=excluded.updated_at_ns`

const (
	currentPlaybackReportSQL = `excluded.played_ms >= playback_metrics.played_ms AND excluded.stalls >= playback_metrics.stalls AND excluded.stall_ms >= playback_metrics.stall_ms`
	mergeDroppedFramesSQL    = `nullif(max(coalesce(playback_metrics.dropped_frames, -1), coalesce(excluded.dropped_frames, -1)), -1)`
	mergeTotalFramesSQL      = `nullif(max(coalesce(playback_metrics.total_frames, -1), coalesce(excluded.total_frames, -1)), -1)`
)

// RecordTranscodeSpeed stores the latest speed (multiple of real time) for the
// user/stream pair, independently of when browser measurements arrive.
func (s *Store) RecordTranscodeSpeed(ctx context.Context, userID int64, streamID string, speed float64) error {
	if !playmetrics.ValidID(streamID) || !playmetrics.ValidSpeed(speed) {
		return fmt.Errorf("invalid transcode speed measurement")
	}
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	now := time.Now()
	if err := expirePlaybackMetrics(ctx, tx, now); err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, `INSERT INTO playback_transcode_speeds(user_id,stream_id,speed,updated_at_ns) VALUES(?,?,?,?)
 ON CONFLICT(user_id,stream_id) DO UPDATE SET speed=excluded.speed, updated_at_ns=excluded.updated_at_ns`,
		userID, strings.ToLower(streamID), speed, now.UnixNano())
	if err != nil {
		return err
	}
	if err := capPlaybackMetrics(ctx, tx); err != nil {
		return err
	}
	return tx.Commit()
}

func (s *Store) ListPlaybackMetrics(ctx context.Context) ([]playmetrics.Item, error) {
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	if err := expirePlaybackMetrics(ctx, tx, time.Now()); err != nil {
		return nil, err
	}
	if err := capPlaybackMetrics(ctx, tx); err != nil {
		return nil, err
	}
	rows, err := tx.QueryContext(ctx, `SELECT m.id,m.mode,m.quality,m.burn_subtitle,m.width,m.height,m.startup_ms,m.stalls,m.stall_ms,
 m.played_ms,m.buffer_ms,m.dropped_frames,m.total_frames,m.updated_at_ns,t.speed,m.source_json,m.reason
 FROM playback_metrics m LEFT JOIN playback_transcode_speeds t
 ON m.user_id=t.user_id AND m.stream_id=t.stream_id AND m.mode='transcode'
 ORDER BY m.updated_at_ns DESC,m.rowid DESC LIMIT ?`, playbackMetricsListLimit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]playmetrics.Item, 0)
	for rows.Next() {
		var item playmetrics.Item
		var updatedNS int64
		var sourceJSON sql.NullString
		if err := rows.Scan(&item.ID, &item.Mode, &item.Quality, &item.BurnSubtitle, &item.Width, &item.Height, &item.StartupMS,
			&item.Stalls, &item.StallMS, &item.PlayedMS, &item.BufferMS, &item.DroppedFrames, &item.TotalFrames,
			&updatedNS, &item.TranscodeSpeed, &sourceJSON, &item.Reason); err != nil {
			return nil, err
		}
		if sourceJSON.Valid {
			if err := json.Unmarshal([]byte(sourceJSON.String), &item.Source); err != nil {
				return nil, fmt.Errorf("decode playback metrics source: %w", err)
			}
		}
		item.UpdatedAt = time.Unix(0, updatedNS).Unix()
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, tx.Commit()
}

var playbackMetricsTables = [...]string{"playback_metrics", "playback_transcode_speeds"}

func expirePlaybackMetrics(ctx context.Context, tx *sql.Tx, now time.Time) error {
	for _, table := range playbackMetricsTables {
		query := `DELETE FROM ` + table + ` WHERE updated_at_ns <= ?`
		if _, err := tx.ExecContext(ctx, query, now.Add(-playbackMetricsRetention).UnixNano()); err != nil {
			return err
		}
	}
	return nil
}

func capPlaybackMetrics(ctx context.Context, tx *sql.Tx) error {
	for _, table := range playbackMetricsTables {
		query := `DELETE FROM ` + table + ` WHERE rowid IN (
 SELECT rowid FROM ` + table + ` ORDER BY updated_at_ns DESC,rowid DESC LIMIT -1 OFFSET ?)`
		if _, err := tx.ExecContext(ctx, query, playbackMetricsMaxRows); err != nil {
			return err
		}
	}
	return nil
}
