package store

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestLibraryPersistenceAndAvailability(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "db.sqlite")
	db, err := Open(ctx, path)
	if err != nil {
		t.Fatal(err)
	}
	root := t.TempDir()
	original, err := db.SaveLibrary(ctx, Library{Name: "电影", Type: "movie", RootPath: root})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	db, err = Open(ctx, path)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	items, err := db.Libraries(ctx)
	if err != nil || len(items) != 1 || items[0].ID != original.ID || items[0].RootPath != root {
		t.Fatalf("persistence failure %+v %v", items, err)
	}
	// Removing an empty source directory must not silently delete its configuration.
	if err := os.Remove(root); err != nil {
		t.Fatal(err)
	}
	items, err = db.Libraries(ctx)
	if err != nil || len(items) != 1 {
		t.Fatal("missing directory removed configuration")
	}
}
