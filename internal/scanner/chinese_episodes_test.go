package scanner

import (
	"context"
	"fmt"
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

func TestChineseEpisodesCreateSeparateSeriesWithoutRenaming(t *testing.T) {
	ctx := context.Background()
	db, err := store.Open(ctx, filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	root := t.TempDir()
	for episode := 1; episode <= 12; episode++ {
		paths := []string{
			fmt.Sprintf("百妖谱.2022.S03/百妖谱.2022.S03E%02d.mp4", episode),
			fmt.Sprintf("百妖谱·司府篇/[百妖谱 司府篇][ 第%02d集 故事][2024][2160P] - Group.mp4", episode),
		}
		for _, path := range paths {
			full := filepath.Join(root, path)
			if err := os.MkdirAll(filepath.Dir(full), 0700); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(full, []byte("broken video"), 0600); err != nil {
				t.Fatal(err)
			}
		}
	}
	lib, err := db.SaveLibrary(ctx, library.Library{Name: "动漫", Type: "tv", RootPath: root})
	if err != nil {
		t.Fatal(err)
	}
	manager := New(db, config.Config{DataDir: t.TempDir()}, slog.New(slog.NewTextHandler(io.Discard, nil)))
	defer manager.Close()
	for run := 0; run < 2; run++ {
		if _, err := manager.Scan(ctx, lib.ID); err != nil {
			t.Fatal(err)
		}
		waitJob(t, manager, lib.ID)
		page, err := db.SeriesPage(ctx, media.Filter{Page: 1, PageSize: 24}, 0)
		if err != nil || page.Total != 2 {
			t.Fatalf("separate chapter series missing: %+v %v", page, err)
		}
		for _, item := range page.Items {
			if item.EpisodeCount != 12 {
				t.Fatalf("lost or duplicated episodes: %+v", item)
			}
		}
	}
}
