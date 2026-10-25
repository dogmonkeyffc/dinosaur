package scanner

import (
	"context"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"dinosaur/internal/config"
	"dinosaur/internal/library"
	"dinosaur/internal/media"
	"dinosaur/internal/store"
)

func TestTVScanStartupIncrementalAndOutage(t *testing.T) {
	ctx := context.Background()
	db, err := store.Open(ctx, filepath.Join(t.TempDir(), "tv.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	root := t.TempDir()
	paths := []string{"Example (2020)/Season 1/Example.S01E01.mp4", "Example (2020)/Season 2/Example.S02E01.mp4", "unknown.mp4"}
	for _, path := range paths {
		full := filepath.Join(root, path)
		if err := os.MkdirAll(filepath.Dir(full), 0700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte("broken video"), 0600); err != nil {
			t.Fatal(err)
		}
	}
	lib, err := db.SaveLibrary(ctx, library.Library{Name: "Series", Type: "tv", RootPath: root})
	if err != nil {
		t.Fatal(err)
	}
	manager := New(db, config.Config{DataDir: t.TempDir()}, slog.New(slog.NewTextHandler(io.Discard, nil)))
	defer manager.Close()
	if err := manager.Start(); err != nil {
		t.Fatal(err)
	}
	job := waitJob(t, manager, lib.ID)
	if job.Status != "completed" || job.Discovered != 3 || job.Processed != 3 || job.Errors != 3 {
		t.Fatalf("job=%+v", job)
	}
	page, err := db.SeriesPage(ctx, media.Filter{Page: 1, PageSize: 24}, 0)
	if err != nil || page.Total != 1 || page.Items[0].EpisodeCount != 2 {
		t.Fatalf("series=%+v %v", page, err)
	}
	episodes, err := db.SeriesEpisodes(ctx, page.Items[0].ID, 0)
	if err != nil || len(episodes) != 2 || !episodes[0].ProbeError {
		t.Fatal("damaged episodes lost")
	}
	if _, err := manager.Scan(ctx, lib.ID); err != nil {
		t.Fatal(err)
	}
	waitJob(t, manager, lib.ID)
	again, err := db.SeriesEpisodes(ctx, page.Items[0].ID, 0)
	if err != nil || again[0].ID != episodes[0].ID {
		t.Fatal("repeat scan changed episode identity")
	}
	if err := os.Rename(root, root+"-offline"); err != nil {
		t.Fatal(err)
	}
	if _, err := manager.Scan(ctx, lib.ID); err != nil {
		t.Fatal(err)
	}
	if job := waitJob(t, manager, lib.ID); job.Status != "failed" {
		t.Fatal("outage accepted")
	}
	if _, err := db.MediaByID(ctx, episodes[0].ID); err != nil {
		t.Fatal("outage hid episodes")
	}
	if err := os.Rename(root+"-offline", root); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(filepath.Join(root, paths[0])); err != nil {
		t.Fatal(err)
	}
	if _, err := manager.Scan(ctx, lib.ID); err != nil {
		t.Fatal(err)
	}
	waitJob(t, manager, lib.ID)
	again, err = db.SeriesEpisodes(ctx, page.Items[0].ID, 0)
	if err != nil || len(again) != 1 || again[0].SeasonNumber != 2 {
		t.Fatal("missing episode not hidden")
	}
}
