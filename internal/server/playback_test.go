package server

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"dinosaur/internal/media"
	"dinosaur/internal/store"
)

func playbackFixture(t *testing.T) (*Server, *store.Store, session, string) {
	t.Helper()
	s, db := fixture(t)
	setupAdmin(t, s)
	admin := loginAs(t, s, "admin", testPassword)
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "Movie.mp4"), []byte("0123456789"), 0600); err != nil {
		t.Fatal(err)
	}
	library, err := db.SaveLibrary(context.Background(), store.Library{Name: "Movies", Type: "movie", RootPath: root})
	if err != nil {
		t.Fatal(err)
	}
	item := media.Item{LibraryID: library.ID, RelativePath: "Movie.mp4", Title: "Movie", Duration: 1000, VideoCodec: "h264", Size: 10, Tracks: []media.Track{{Index: 1, Type: "audio", Codec: "aac"}}}
	if err := db.SaveMedia(context.Background(), item, 1); err != nil {
		t.Fatal(err)
	}
	return s, db, admin, root
}
func TestPlaybackRangeAndSourceProtection(t *testing.T) {
	s, _, admin, root := playbackFixture(t)
	path := "/api/playback/1/stream?mode=original"
	status(t, request(t, s, "GET", path, nil, nil, nil), 401)
	w := request(t, s, "GET", path, nil, &admin, map[string]string{"Range": "bytes=2-5"})
	if w.Code != http.StatusPartialContent || w.Body.String() != "2345" || w.Header().Get("Content-Range") != "bytes 2-5/10" {
		t.Fatalf("invalid range response %d %s %v", w.Code, w.Body.String(), w.Header())
	}
	if w.Header().Get("Cache-Control") != "no-store" {
		t.Fatal("protected media may be cached")
	}
	w = request(t, s, "HEAD", path, nil, &admin, nil)
	if w.Code != 200 || w.Body.Len() != 0 || w.Header().Get("Content-Length") != "10" {
		t.Fatal("invalid HEAD")
	}
	w = request(t, s, "GET", path, nil, &admin, map[string]string{"Range": "bytes=50-60"})
	if w.Code != 416 {
		t.Fatalf("unsatisfiable range=%d", w.Code)
	}
	status(t, request(t, s, "GET", "/api/playback/1?start=NaN", nil, &admin, nil), 400)
	status(t, request(t, s, "GET", "/api/playback/1?audio=999", nil, &admin, nil), 400)
	outside := filepath.Join(t.TempDir(), "secret.mp4")
	if err := os.WriteFile(outside, []byte("private"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(filepath.Join(root, "Movie.mp4")); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, filepath.Join(root, "Movie.mp4")); err != nil {
		t.Fatal(err)
	}
	status(t, request(t, s, "GET", path, nil, &admin, nil), 404)
}
func TestProgressIsolationAndContinueWatching(t *testing.T) {
	s, db, admin, _ := playbackFixture(t)
	body := map[string]any{"position": 125}
	status(t, request(t, s, "PUT", "/api/playback/1/progress", body, &admin, map[string]string{"X-CSRF-Token": ""}), 403)
	status(t, request(t, s, "PUT", "/api/playback/1/progress", body, &admin, nil), 200)
	first := request(t, s, "GET", "/api/playback/1", nil, &admin, nil)
	status(t, first, 200)
	if !strings.Contains(first.Body.String(), `"resume_position":125`) {
		t.Fatal("resume not returned")
	}
	user, err := db.UserByName(context.Background(), "admin")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.CreateUser(context.Background(), "viewer", user.PasswordHash, "user", false); err != nil {
		t.Fatal(err)
	}
	viewer := loginAs(t, s, "viewer", testPassword)
	other := request(t, s, "GET", "/api/playback/1", nil, &viewer, nil)
	status(t, other, 200)
	if !strings.Contains(other.Body.String(), `"resume_position":0`) {
		t.Fatal("cross-user progress exposed")
	}
	for _, test := range []struct {
		who   session
		count int
	}{{admin, 1}, {viewer, 0}} {
		w := request(t, s, "GET", "/api/playback/continue", nil, &test.who, nil)
		status(t, w, 200)
		var result struct {
			Items []media.Item `json:"items"`
		}
		if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
			t.Fatal(err)
		}
		if len(result.Items) != test.count {
			t.Fatalf("continue count=%d", len(result.Items))
		}
	}
	status(t, request(t, s, "PUT", "/api/playback/1/progress", map[string]any{"watched": true}, &admin, nil), 200)
	state, err := db.PlaybackProgress(context.Background(), 1, 1)
	if err != nil || !state.Watched || state.Position != 1000 {
		t.Fatalf("completion=%+v %v", state, err)
	}
	status(t, request(t, s, "PUT", "/api/playback/1/progress", map[string]any{"watched": false}, &admin, nil), 200)
	status(t, request(t, s, "PUT", "/api/playback/1/progress", map[string]any{"position": -1}, &admin, nil), 400)
	status(t, request(t, s, "PUT", "/api/playback/1/progress", map[string]any{}, &admin, nil), 400)
}
