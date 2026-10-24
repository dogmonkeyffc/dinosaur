package server

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"dinosaur/internal/store"
)

func TestLibraryConfiguration(t *testing.T) {
	s, db := fixture(t)
	setupAdmin(t, s)
	admin := loginAs(t, s, "admin", testPassword)
	root := t.TempDir()
	mediaFile := filepath.Join(root, "keep.mp4")
	if err := os.WriteFile(mediaFile, []byte("original media"), 0600); err != nil {
		t.Fatal(err)
	}
	body := libraryRequest{Name: "电影收藏", Type: "movie", RootPath: root}
	status(t, request(t, s, "POST", "/api/admin/libraries", body, nil, nil), 401)
	status(t, request(t, s, "POST", "/api/admin/libraries", body, &admin, map[string]string{"X-CSRF-Token": ""}), 403)
	created := request(t, s, "POST", "/api/admin/libraries", body, &admin, nil)
	status(t, created, 201)
	var result struct {
		Library store.Library `json:"library"`
	}
	if err := json.Unmarshal(created.Body.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	if result.Library.ID != 1 || result.Library.RootPath != root {
		t.Fatalf("unexpected library %+v", result.Library)
	}
	status(t, request(t, s, "POST", "/api/admin/libraries", body, &admin, nil), 409)
	alias := filepath.Join(t.TempDir(), "alias")
	if err := os.Symlink(root, alias); err != nil {
		t.Fatal(err)
	}
	body.RootPath = alias
	status(t, request(t, s, "POST", "/api/admin/libraries", body, &admin, nil), 409)
	for _, bad := range []libraryRequest{
		{Name: "", Type: "movie", RootPath: root},
		{Name: "Bad", Type: "music", RootPath: root},
		{Name: "Bad", Type: "movie", RootPath: "relative/path"},
		{Name: "Bad", Type: "movie", RootPath: filepath.Join(root, "missing")},
		{Name: "Bad", Type: "movie", RootPath: mediaFile},
	} {
		status(t, request(t, s, "POST", "/api/admin/libraries", bad, &admin, nil), 400)
	}
	waitForLibraryScan(t, s, 1)
	body = libraryRequest{Name: "剧集收藏", Type: "tv", RootPath: root}
	status(t, request(t, s, "PUT", "/api/admin/libraries/1", body, &admin, nil), 200)
	list := request(t, s, "GET", "/api/admin/libraries", nil, &admin, nil)
	status(t, list, 200)
	if !strings.Contains(list.Body.String(), `"available":true`) || !strings.Contains(list.Body.String(), "剧集收藏") {
		t.Fatalf("bad list %s", list.Body.String())
	}
	// The authenticated public list must not reveal server filesystem paths.
	_, err := db.CreateUser(context.Background(), "viewer", s.dummyHash, "user", false)
	if err != nil {
		t.Fatal(err)
	}
	hash, err := s.hasher.Hash(context.Background(), testPassword)
	if err != nil {
		t.Fatal(err)
	}
	u, err := db.UserByName(context.Background(), "viewer")
	if err != nil {
		t.Fatal(err)
	}
	if err := db.ChangePassword(context.Background(), u.ID, hash); err != nil {
		t.Fatal(err)
	}
	viewer := loginAs(t, s, "viewer", testPassword)
	public := request(t, s, "GET", "/api/libraries", nil, &viewer, nil)
	status(t, public, 200)
	if strings.Contains(public.Body.String(), root) || strings.Contains(public.Body.String(), "root_path") {
		t.Fatal("server directory leaked")
	}
	for _, method := range []string{"GET", "POST", "PUT", "DELETE"} {
		path := "/api/admin/libraries"
		if method == "PUT" || method == "DELETE" {
			path += "/1"
		}
		status(t, request(t, s, method, path, body, &viewer, nil), 403)
	}
	status(t, request(t, s, "PUT", "/api/admin/libraries/999", body, &admin, nil), 404)
	status(t, request(t, s, "DELETE", "/api/admin/libraries/invalid", nil, &admin, nil), 400)
	status(t, request(t, s, "DELETE", "/api/admin/libraries/1", nil, &admin, nil), 200)
	if data, err := os.ReadFile(mediaFile); err != nil || string(data) != "original media" {
		t.Fatal("media file changed or removed")
	}
	status(t, request(t, s, "DELETE", "/api/admin/libraries/1", nil, &admin, nil), 404)
	items, err := db.Libraries(context.Background())
	if err != nil || len(items) != 0 {
		t.Fatalf("libraries not removed: %v", err)
	}
}
func TestEmbeddedWeb(t *testing.T) {
	s, _ := fixture(t)
	for _, path := range []string{"/", "/login", "/setup", "/settings", "/watch/1"} {
		w := request(t, s, "GET", path, nil, nil, nil)
		if w.Code != 200 || !strings.HasPrefix(w.Header().Get("Content-Type"), "text/html") || !strings.Contains(w.Body.String(), "Dinosaur") {
			t.Fatalf("bad web response %s: %d", path, w.Code)
		}
		if !strings.Contains(w.Header().Get("Content-Security-Policy"), "frame-ancestors 'none'") {
			t.Fatal("CSP missing")
		}
	}
	for path, kind := range map[string]string{"/assets/app.css": "text/css", "/assets/app.js": "text/javascript", "/assets/vendor/artplayer.js": "text/javascript", "/assets/vendor/artplayer.css": "text/css"} {
		w := request(t, s, "GET", path, nil, nil, nil)
		if w.Code != 200 || !strings.HasPrefix(w.Header().Get("Content-Type"), kind) {
			t.Fatalf("bad asset %s", path)
		}
	}
	status(t, request(t, s, "GET", "/assets/missing.js", nil, nil, nil), 404)
}
