package store

import (
	"context"
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"dinosaur/internal/playmetrics"
)

func playbackMetricsFixture(t *testing.T) (*Store, int64) {
	t.Helper()
	ctx := context.Background()
	db, err := Open(ctx, filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	u, err := db.CreateUser(ctx, "admin", "hash", "admin", true)
	if err != nil {
		t.Fatal(err)
	}
	return db, u.ID
}

func metricsTestID(n int) string { return fmt.Sprintf("00000000-0000-4000-8000-%012x", n) }

func TestPlaybackMetricsCumulativeUpdate(t *testing.T) {
	ctx := context.Background()
	db, userID := playbackMetricsFixture(t)
	streamID := metricsTestID(2)
	startup, dropped, total, subtitle := int64(500), int64(5), int64(100), false
	snapshot := playmetrics.Snapshot{StreamID: &streamID, Metrics: playmetrics.Metrics{
		ID: metricsTestID(1), Mode: "transcode", Quality: "720", Width: 1280, Height: 720,
		StartupMS: &startup, Stalls: 2, StallMS: 200, PlayedMS: 1200, BufferMS: 3000,
		DroppedFrames: &dropped, TotalFrames: &total, BurnSubtitle: &subtitle,
	}}
	if err := db.SavePlaybackMetrics(ctx, userID, snapshot); err != nil {
		t.Fatal(err)
	}
	if err := db.RecordTranscodeSpeed(ctx, userID, streamID, 0.5); err != nil {
		t.Fatal(err)
	}
	if err := db.RecordTranscodeSpeed(ctx, userID, streamID, 1.5); err != nil {
		t.Fatal(err)
	}
	older := playmetrics.Snapshot{Metrics: playmetrics.Metrics{
		ID: snapshot.ID, Mode: "direct", Quality: "auto", Width: 640, Height: 480,
		Stalls: 1, StallMS: 100, PlayedMS: 1000, BufferMS: 8000,
	}}
	if err := db.SavePlaybackMetrics(ctx, userID, older); err != nil {
		t.Fatal(err)
	}
	items := readMetricItems(t, db)
	if len(items) != 1 || items[0].Mode != "transcode" || items[0].BufferMS != 3000 || items[0].Stalls != 2 || items[0].StallMS != 200 || items[0].PlayedMS != 1200 {
		t.Fatalf("stale update overwrote snapshot: %+v", items)
	}

	newer := snapshot
	newer.PlayedMS, newer.BufferMS, newer.Width, newer.Height = 1800, 1500, 0, 0
	newer.StartupMS, newer.BurnSubtitle = nil, nil
	newer.DroppedFrames, newer.TotalFrames, newer.StreamID = nil, nil, nil
	if err := db.SavePlaybackMetrics(ctx, userID, newer); err != nil {
		t.Fatal(err)
	}
	items = readMetricItems(t, db)
	if len(items) != 1 {
		t.Fatalf("got %d items", len(items))
	}
	item := items[0]
	if item.StartupMS == nil || *item.StartupMS != 500 || item.DroppedFrames == nil || *item.DroppedFrames != 5 || item.TotalFrames == nil || *item.TotalFrames != 100 || item.BurnSubtitle == nil || *item.BurnSubtitle {
		t.Fatalf("known optional fields lost: %+v", item)
	}
	if item.PlayedMS != 1800 || item.BufferMS != 1500 || item.Width != 1280 || item.TranscodeSpeed == nil || *item.TranscodeSpeed != 1.5 {
		t.Fatalf("unexpected latest snapshot: %+v", item)
	}
}

func TestPlaybackMetricsExpiryAndLimits(t *testing.T) {
	ctx := context.Background()
	db, userID := playbackMetricsFixture(t)
	now := time.Now()
	expired := now.Add(-playbackMetricsRetention - time.Second).UnixNano()
	tx, err := db.DB.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback()
	for i := 0; i <= playbackMetricsMaxRows; i++ {
		id := metricsTestID(i)
		updated := now.Add(-time.Minute + time.Duration(i)*time.Nanosecond).UnixNano()
		if i == 0 {
			updated = expired
		}
		if _, err := tx.Exec(`INSERT INTO playback_metrics(user_id,id,stream_id,mode,quality,width,height,stalls,stall_ms,played_ms,buffer_ms,updated_at_ns)
 VALUES(?,?,?,'transcode','auto',0,0,0,0,0,0,?)`, userID, id, id, updated); err != nil {
			t.Fatal(err)
		}
		if _, err := tx.Exec(`INSERT INTO playback_transcode_speeds(user_id,stream_id,speed,updated_at_ns) VALUES(?,?,1.5,?)`, userID, id, updated); err != nil {
			t.Fatal(err)
		}
	}
	if err := tx.Commit(); err != nil {
		t.Fatal(err)
	}
	// A speed write expires rows from both tables, and caps anonymous orphan speed rows.
	if err := db.RecordTranscodeSpeed(ctx, userID, metricsTestID(playbackMetricsMaxRows+1), 2); err != nil {
		t.Fatal(err)
	}
	assertMetricTable(t, db, "playback_metrics", "id", 1)
	assertMetricTable(t, db, "playback_transcode_speeds", "stream_id", 2)
	newest := playmetrics.Snapshot{Metrics: playmetrics.Metrics{ID: metricsTestID(playbackMetricsMaxRows + 2), Mode: "direct", Quality: "auto"}}
	if err := db.SavePlaybackMetrics(ctx, userID, newest); err != nil {
		t.Fatal(err)
	}
	assertMetricTable(t, db, "playback_metrics", "id", 2)

	// Expire a speed with a live browser measurement, and a browser row with live speed.
	if _, err := db.DB.Exec(`UPDATE playback_transcode_speeds SET updated_at_ns=? WHERE stream_id=?`, expired, metricsTestID(playbackMetricsMaxRows)); err != nil {
		t.Fatal(err)
	}
	if _, err := db.DB.Exec(`UPDATE playback_metrics SET updated_at_ns=? WHERE id=?`, expired, metricsTestID(playbackMetricsMaxRows-1)); err != nil {
		t.Fatal(err)
	}
	items := readMetricItems(t, db)
	if len(items) != playbackMetricsListLimit || items[0].ID != newest.ID || items[1].ID != metricsTestID(playbackMetricsMaxRows) || items[1].TranscodeSpeed != nil || items[2].ID != metricsTestID(playbackMetricsMaxRows-2) {
		t.Fatalf("unexpected list length/order or stale join: %d, %.300v", len(items), items)
	}
	for _, table := range playbackMetricsTables {
		var count int
		if err := db.DB.QueryRow(`SELECT count(*) FROM `+table+` WHERE updated_at_ns <= ?`, expired).Scan(&count); err != nil || count != 0 {
			t.Fatalf("%s expired rows=%d err=%v", table, count, err)
		}
	}
}

func readMetricItems(t *testing.T, db *Store) []playmetrics.Item {
	t.Helper()
	items, err := db.ListPlaybackMetrics(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	return items
}

func assertMetricTable(t *testing.T, db *Store, table, key string, oldest int) {
	t.Helper()
	var count int
	var id string
	if err := db.DB.QueryRow(`SELECT count(*),min(`+key+`) FROM `+table).Scan(&count, &id); err != nil || count != playbackMetricsMaxRows || id != metricsTestID(oldest) {
		t.Fatalf("%s count=%d min=%q err=%v", table, count, id, err)
	}
}
