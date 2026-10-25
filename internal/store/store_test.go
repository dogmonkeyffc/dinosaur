package store

import (
	"context"
	"errors"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

func TestPersistenceAndSQLiteConstraints(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "data", "dinosaur.db")
	db, err := Open(ctx, path)
	if err != nil {
		t.Fatal(err)
	}
	for key, want := range map[string]int{"foreign_keys": 1, "busy_timeout": 5000} {
		var got int
		if err := db.DB.QueryRow("PRAGMA " + key).Scan(&got); err != nil {
			t.Fatal(err)
		}
		if got != want {
			t.Fatalf("%s=%d", key, got)
		}
	}
	var mode string
	if err := db.DB.QueryRow("PRAGMA journal_mode").Scan(&mode); err != nil {
		t.Fatal(err)
	}
	if mode != "wal" {
		t.Fatalf("mode=%s", mode)
	}
	u, err := db.CreateUser(ctx, "admin", "hash", "admin", true)
	if err != nil {
		t.Fatal(err)
	}
	if err := db.CreateSession(ctx, "token", u.ID, "hash", time.Now().Add(time.Hour)); err != nil {
		t.Fatal(err)
	}
	if err := db.CreateSession(ctx, "stale", u.ID, "old-hash", time.Now().Add(time.Hour)); err == nil {
		t.Fatal("login with stale password hash succeeded")
	}
	if _, err := db.DB.Exec(`INSERT INTO sessions VALUES ('invalid',99999,9999999999,0)`); err == nil {
		t.Fatal("foreign key not enforced")
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	db, err = Open(ctx, path)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	stored, err := db.SessionUser(ctx, "token")
	if err != nil || stored.ID != u.ID {
		t.Fatalf("session not persisted: %v", err)
	}
	if _, err := db.CreateUser(ctx, "ADMIN", "hash", "user", false); !errors.Is(err, ErrConflict) {
		t.Fatalf("duplicate error=%v", err)
	}
	var count int
	if err := db.DB.QueryRow(`SELECT count(*) FROM schema_migrations`).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 8 {
		t.Fatalf("migrations=%d", count)
	}
	if err := db.ChangePassword(ctx, u.ID, "new-hash"); err != nil {
		t.Fatal(err)
	}
	if _, err := db.SessionUser(ctx, "token"); err == nil {
		t.Fatal("session survived password change")
	}
}
func TestConcurrentInitializationAcrossConnections(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "test.db")
	a, err := Open(ctx, path)
	if err != nil {
		t.Fatal(err)
	}
	defer a.Close()
	b, err := Open(ctx, path)
	if err != nil {
		t.Fatal(err)
	}
	defer b.Close()
	results := make(chan error, 2)
	var wg sync.WaitGroup
	for _, db := range []*Store{a, b} {
		wg.Add(1)
		go func(db *Store) {
			defer wg.Done()
			_, err := db.CreateUser(ctx, "admin", "hash", "admin", true)
			results <- err
		}(db)
	}
	wg.Wait()
	close(results)
	success := 0
	for err := range results {
		if err == nil {
			success++
		} else if !errors.Is(err, ErrInitialized) {
			t.Fatalf("unexpected error=%v", err)
		}
	}
	if success != 1 {
		t.Fatalf("created %d", success)
	}
}
