package scanner

import (
	"context"
	"io"
	"log/slog"
	"os/exec"
	"path/filepath"
	"testing"

	"dinosaur/internal/config"
	"dinosaur/internal/library"
	"dinosaur/internal/media"
	"dinosaur/internal/store"
)

func TestOrdinaryVideosKeepFilenamesWithoutSeriesGrouping(t *testing.T) {
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		t.Skip("ffmpeg unavailable")
	}
	ctx := context.Background()
	db, err := store.Open(ctx, filepath.Join(t.TempDir(), "video.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	root := t.TempDir()
	name := "Holiday.2024.S01E01.mp4"
	cmd := exec.Command("ffmpeg", "-v", "error", "-f", "lavfi", "-i", "color=c=blue:s=160x120:d=1", "-c:v", "mpeg4", "-threads", "1", filepath.Join(root, name))
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("fixture: %s %v", output, err)
	}
	lib, err := library.Validate(library.Library{Name: "Videos", Type: "video", RootPath: root})
	if err != nil {
		t.Fatal(err)
	}
	lib, err = db.SaveLibrary(ctx, lib)
	if err != nil {
		t.Fatal(err)
	}
	m := New(db, config.Config{DataDir: t.TempDir()}, slog.New(slog.NewTextHandler(io.Discard, nil)))
	defer m.Close()
	var id int64
	for range 2 {
		if _, err := m.Scan(ctx, lib.ID); err != nil {
			t.Fatal(err)
		}
		if job := waitJob(t, m, lib.ID); job.Status != "completed" || job.Errors != 0 {
			t.Fatalf("scan: %+v", job)
		}
		page, err := db.MediaPage(ctx, media.Filter{Type: "video", Page: 1, PageSize: 24})
		if err != nil || page.Total != 1 {
			t.Fatalf("catalog: %+v %v", page, err)
		}
		item := page.Items[0]
		if item.Title != name || item.Year != 0 || item.SeriesID != 0 || item.Metadata != nil {
			t.Fatalf("video misidentified: %+v", item)
		}
		if id != 0 && item.ID != id {
			t.Fatal("incremental scan changed identity")
		}
		id = item.ID
	}
	series, err := db.SeriesPage(ctx, media.Filter{Page: 1, PageSize: 24}, 0)
	if err != nil || series.Total != 0 {
		t.Fatal("video grouped as a series", err)
	}
}
