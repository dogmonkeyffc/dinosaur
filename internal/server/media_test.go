package server

import (
	"context"
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"dinosaur/internal/media"
)

func waitForLibraryScan(t *testing.T, s *Server, id int64) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		job, err := s.store.LatestScan(context.Background(), id)
		if err != nil {
			t.Fatal(err)
		}
		if job != nil && job.Status != "queued" && job.Status != "scanning" && s.scanner.WithIdle(id, func() error { return nil }) == nil {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("scan timeout")
}
func TestMediaAccessSearchAndPagination(t *testing.T) {
	s, db := fixture(t)
	setupAdmin(t, s)
	admin := loginAs(t, s, "admin", testPassword)
	root := t.TempDir()
	created := request(t, s, "POST", "/api/admin/libraries", libraryRequest{Name: "电影", Type: "movie", RootPath: root}, &admin, nil)
	status(t, created, 201)
	waitForLibraryScan(t, s, 1)
	for _, title := range []string{"Example", "Other", "Literal%Title"} {
		item := media.Item{LibraryID: 1, RelativePath: title + ".mkv", Title: title, Year: 2024, Size: 100, Tracks: make([]media.Track, 0)}
		if err := db.SaveMedia(context.Background(), item, 1); err != nil {
			t.Fatal(err)
		}
	}
	status(t, request(t, s, "GET", "/api/media", nil, nil, nil), 401)
	status(t, request(t, s, "GET", "/api/media/1/poster", nil, nil, nil), 401)
	for query, want := range map[string]int{"?page=0": 400, "?page_size=101": 400, "?library_id=-1": 400, "?q=Example": 200, "?q=%25": 200, "?page=2&page_size=1": 200} {
		w := request(t, s, "GET", "/api/media"+query, nil, &admin, nil)
		status(t, w, want)
		if want == 200 {
			var page media.Page
			if err := json.Unmarshal(w.Body.Bytes(), &page); err != nil {
				t.Fatal(err)
			}
			if len(page.Items) != 1 {
				t.Fatalf("page=%+v", page)
			}
			if strings.Contains(w.Body.String(), root) || strings.Contains(w.Body.String(), "relative_path") {
				t.Fatal("media path leaked")
			}
		}
	}
	status(t, request(t, s, "GET", "/api/media/1", nil, &admin, nil), 200)
	status(t, request(t, s, "GET", "/api/media/1/poster", nil, &admin, nil), 404)
	status(t, request(t, s, "GET", "/api/media/999", nil, &admin, nil), 404)
	// Changing the root invalidates the previous index, without touching files.
	updated := libraryRequest{Name: "New root", Type: "movie", RootPath: filepath.Join(t.TempDir())}
	status(t, request(t, s, "PUT", "/api/admin/libraries/1", updated, &admin, nil), 200)
	status(t, request(t, s, "GET", "/api/media/1", nil, &admin, nil), 404)
}
