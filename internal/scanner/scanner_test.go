package scanner

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"dinosaur/internal/config"
	"dinosaur/internal/library"
	"dinosaur/internal/media"
	"dinosaur/internal/store"
)

func waitJob(t *testing.T, m *Manager, id int64) *library.ScanJob {
	t.Helper()
	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		job, err := m.db.LatestScan(context.Background(), id)
		if err != nil {
			t.Fatal(err)
		}
		if job != nil && job.Status != "queued" && job.Status != "scanning" && m.WithIdle(id, func() error { return nil }) == nil {
			return job
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("scan did not finish")
	return nil
}
func TestScanIncrementalAndMissing(t *testing.T) {
	for _, tool := range []string{"ffmpeg", "ffprobe"} {
		if _, err := exec.LookPath(tool); err != nil {
			t.Skip("media tools not installed")
		}
	}
	ctx := context.Background()
	db, err := store.Open(ctx, filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	root := t.TempDir()
	good := filepath.Join(root, "Example.2024.mkv")
	cmd := exec.Command("ffmpeg", "-v", "error", "-f", "lavfi", "-i", "color=c=green:s=160x120:d=1", "-c:v", "mpeg4", "-threads", "1", good)
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("fixture: %s %v", output, err)
	}
	broken := filepath.Join(root, "Broken.2023.mkv")
	if err := os.WriteFile(broken, []byte("invalid"), 0600); err != nil {
		t.Fatal(err)
	}
	outside := filepath.Join(t.TempDir(), "ignored.mkv")
	if err := os.WriteFile(outside, []byte("outside"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, filepath.Join(root, "escape.mkv")); err != nil {
		t.Fatal(err)
	}
	item, err := db.SaveLibrary(ctx, library.Library{Name: "Movies", Type: "movie", RootPath: root})
	if err != nil {
		t.Fatal(err)
	}
	m := New(db, config.Config{DataDir: t.TempDir()}, slog.New(slog.NewTextHandler(io.Discard, nil)))
	defer m.Close()
	// Occupy both worker slots to make queue and duplicate checks deterministic.
	m.slots <- struct{}{}
	m.slots <- struct{}{}
	if _, err := m.Scan(ctx, item.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := m.Scan(ctx, item.ID); !errors.Is(err, ErrRunning) {
		t.Fatalf("duplicate: %v", err)
	}
	if err := m.WithIdle(item.ID, func() error { return nil }); !errors.Is(err, ErrRunning) {
		t.Fatal("mutation permitted during scan")
	}
	<-m.slots
	<-m.slots
	job := waitJob(t, m, item.ID)
	if job.Status != "completed" || job.Discovered != 2 || job.Processed != 2 || job.Errors < 1 {
		t.Fatalf("job=%+v", job)
	}
	page, err := db.MediaPage(ctx, media.Filter{Page: 1, PageSize: 24})
	if err != nil {
		t.Fatal(err)
	}
	if page.Total != 2 {
		t.Fatalf("items=%+v", page)
	}
	var goodItem media.Item
	for _, entry := range page.Items {
		if entry.Title == "Example" {
			goodItem = entry
		}
	}
	if goodItem.Duration <= 0 || goodItem.Width != 160 || goodItem.PosterKey == "" {
		t.Fatalf("metadata=%+v", goodItem)
	}
	if _, err := m.Scan(ctx, item.ID); err != nil {
		t.Fatal(err)
	}
	waitJob(t, m, item.ID)
	again, err := db.MediaByID(ctx, goodItem.ID)
	if err != nil || again.ID != goodItem.ID {
		t.Fatal("rescan changed identity")
	}
	if err := os.Rename(root, root+"-offline"); err != nil {
		t.Fatal(err)
	}
	if _, err := m.Scan(ctx, item.ID); err != nil {
		t.Fatal(err)
	}
	if job := waitJob(t, m, item.ID); job.Status != "failed" {
		t.Fatal("unreadable directory scan succeeded")
	}
	if _, err := db.MediaByID(ctx, goodItem.ID); err != nil {
		t.Fatal("directory outage hid media")
	}
	if err := os.Rename(root+"-offline", root); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(good); err != nil {
		t.Fatal(err)
	}
	if _, err := m.Scan(ctx, item.ID); err != nil {
		t.Fatal(err)
	}
	waitJob(t, m, item.ID)
	if _, err := db.MediaByID(ctx, goodItem.ID); err == nil {
		t.Fatal("missing movie still visible")
	}
}
