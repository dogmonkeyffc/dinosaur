package store

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"
	"time"

	"dinosaur/internal/playmetrics"
)

func TestPlaybackMetricsKeepsFirstKnownSourceAndReason(t *testing.T) {
	ctx := context.Background()
	db, userID := playbackMetricsFixture(t)
	snapshot := playmetrics.Snapshot{Metrics: playmetrics.Metrics{
		ID: metricsTestID(1), Mode: "transcode", Quality: "720", PlayedMS: 1000, Width: 1280, Height: 720,
	}}
	wantSource := playmetrics.Source{
		Title: "Example 电影", Path: "TV/Example 电影.mkv", Width: 3840, Height: 2160,
		VideoCodec: "hevc", AudioCodec: "dts", Container: "matroska", BitrateBPS: 20000000, SizeBytes: 7200000000,
	}
	save := func(s playmetrics.Snapshot) {
		t.Helper()
		if err := db.SavePlaybackMetrics(ctx, userID, s); err != nil {
			t.Fatal(err)
		}
	}
	assertMetadata := func(src *playmetrics.Source, reason string) {
		t.Helper()
		items := readMetricItems(t, db)
		if len(items) != 1 {
			t.Fatalf("got %d items", len(items))
		}
		item := items[0]
		if (item.Source == nil) != (src == nil) || (src != nil && *item.Source != *src) || item.Reason != reason {
			t.Fatalf("metadata = (%+v, %q), want (%+v, %q)", item.Source, item.Reason, src, reason)
		}
		if item.Width != 1280 || item.Height != 720 {
			t.Fatalf("original dimensions replaced output dimensions: %+v", item)
		}
	}

	save(snapshot)
	assertMetadata(nil, "")
	stale := snapshot
	stale.PlayedMS = 100
	stale.Source = &wantSource
	save(stale)
	assertMetadata(&wantSource, "")

	newer := snapshot
	newer.PlayedMS = 2000
	newer.Source = &playmetrics.Source{Title: "Not the original source"}
	newer.Reason = "unsupported_video_codec"
	save(newer)
	assertMetadata(&wantSource, "unsupported_video_codec")

	newer.PlayedMS = 3000
	newer.Reason = "different_reason"
	save(newer)
	newer.Source, newer.Reason = nil, ""
	save(newer)
	assertMetadata(&wantSource, "unsupported_video_codec")
	if got := readMetricItems(t, db)[0].PlayedMS; got != 3000 {
		t.Fatalf("metadata immutability stopped counter updates: played_ms=%d", got)
	}
}

func TestPlaybackMetricsSourceMigrationPreservesExistingRows(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "upgrade.db")
	old, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	defer old.Close()
	if _, err := old.Exec(`CREATE TABLE schema_migrations(version TEXT PRIMARY KEY, applied_at INTEGER NOT NULL)`); err != nil {
		t.Fatal(err)
	}
	names, err := migrations.ReadDir("migrations")
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range names {
		if name.Name() >= "016" {
			break
		}
		content, err := migrations.ReadFile("migrations/" + name.Name())
		if err != nil {
			t.Fatal(err)
		}
		if _, err := old.Exec(string(content)); err != nil {
			t.Fatal(err)
		}
		if _, err := old.Exec(`INSERT INTO schema_migrations VALUES (?,0)`, name.Name()); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := old.Exec(`INSERT INTO users(id,username,password_hash,role,created_at) VALUES (1,'admin','hash','admin',1)`); err != nil {
		t.Fatal(err)
	}
	id := metricsTestID(1)
	if _, err := old.Exec(`INSERT INTO playback_metrics(user_id,id,mode,quality,width,height,stalls,stall_ms,played_ms,buffer_ms,updated_at_ns)
 VALUES(1,?,'direct','original',1920,1080,0,0,250,500,?)`, id, time.Now().UnixNano()); err != nil {
		t.Fatal(err)
	}
	if err := old.Close(); err != nil {
		t.Fatal(err)
	}
	db, err := Open(ctx, path)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	items := readMetricItems(t, db)
	if len(items) != 1 || items[0].ID != id || items[0].PlayedMS != 250 || items[0].Source != nil || items[0].Reason != "" {
		t.Fatalf("legacy row changed: %+v", items)
	}
}
