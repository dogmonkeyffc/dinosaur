package server

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"dinosaur/internal/media"
	"dinosaur/internal/store"
	"dinosaur/internal/television"
)

func TestSeriesBrowsingPlaybackAndPermissions(t *testing.T) {
	s, db := fixture(t)
	setupAdmin(t, s)
	admin := loginAs(t, s, "admin", testPassword)
	ctx := context.Background()
	root := t.TempDir()
	lib, err := db.SaveLibrary(ctx, store.Library{Name: "TV", Type: "tv", RootPath: root})
	if err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{"Example (2020)/Example.S01E01.mp4", "Example (2020)/Example.S02E01.mp4"} {
		if err := os.MkdirAll(filepath.Join(root, filepath.Dir(path)), 0700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(root, path), []byte("0123456789"), 0600); err != nil {
			t.Fatal(err)
		}
		name, _ := television.ParseEpisode(path)
		if err := db.SaveEpisode(ctx, media.Item{LibraryID: lib.ID, RelativePath: path, Title: name.Title, Duration: 100, VideoCodec: "h264"}, name, 1); err != nil {
			t.Fatal(err)
		}
	}
	user, err := db.UserByName(ctx, "admin")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.CreateUser(ctx, "viewer", user.PasswordHash, "user", false); err != nil {
		t.Fatal(err)
	}
	viewer := loginAs(t, s, "viewer", testPassword)
	for _, path := range []string{"/api/series", "/api/series/1", "/api/series/1/poster"} {
		status(t, request(t, s, "GET", path, nil, nil, nil), 401)
	}
	page := request(t, s, "GET", "/api/series", nil, &viewer, nil)
	status(t, page, 200)
	if strings.Contains(page.Body.String(), root) || !strings.Contains(page.Body.String(), `"episode_count":2`) {
		t.Fatal("root leaked or series was not grouped")
	}
	for _, path := range []string{"/api/series?page=0", "/api/series?page_size=101", "/api/series?library_id=-1", "/api/playback/continue?type=unknown"} {
		status(t, request(t, s, "GET", path, nil, &viewer, nil), 400)
	}
	status(t, request(t, s, "GET", "/api/series/999", nil, &viewer, nil), 404)
	status(t, request(t, s, "POST", "/api/admin/tv/metadata/jobs", map[string]any{}, &viewer, nil), 403)
	status(t, request(t, s, "GET", "/api/admin/series/1/metadata/candidates", nil, &viewer, nil), 403)
	status(t, request(t, s, "POST", "/api/admin/tv/metadata/jobs", map[string]any{}, &admin, map[string]string{"X-CSRF-Token": ""}), 403)
	status(t, request(t, s, "POST", "/api/admin/tv/metadata/jobs", map[string]any{"series_id": 1}, &admin, nil), 400)
	episode, err := db.SeriesEpisodes(ctx, 1, user.ID)
	if err != nil {
		t.Fatal(err)
	}
	info := request(t, s, "GET", fmt.Sprintf("/api/playback/%d", episode[0].ID), nil, &viewer, nil)
	status(t, info, 200)
	var data struct {
		Next *media.Item `json:"next_episode"`
	}
	if err := json.Unmarshal(info.Body.Bytes(), &data); err != nil || data.Next == nil || data.Next.SeasonNumber != 2 {
		t.Fatal("playback lost next season")
	}
	stream := request(t, s, "GET", fmt.Sprintf("/api/playback/%d/stream?mode=original", episode[0].ID), nil, &viewer, map[string]string{"Range": "bytes=2-5"})
	if stream.Code != 206 || stream.Body.String() != "2345" {
		t.Fatal("episode range playback failed")
	}
	progressPath := fmt.Sprintf("/api/playback/%d/progress", episode[0].ID)
	status(t, request(t, s, "PUT", progressPath, map[string]any{"position": 35}, &viewer, nil), 200)
	detail := request(t, s, "GET", "/api/series/1", nil, &viewer, nil)
	status(t, detail, 200)
	var series television.Detail
	if err := json.Unmarshal(detail.Body.Bytes(), &series); err != nil || len(series.Seasons) != 2 || series.Resume.Position != 35 {
		t.Fatal("episode progress not reflected in detail")
	}
	other := request(t, s, "GET", "/api/series/1", nil, &admin, nil)
	status(t, other, 200)
	if strings.Contains(other.Body.String(), `"position":35`) {
		t.Fatal("episode progress leaked to another user")
	}
	status(t, request(t, s, "PUT", progressPath, map[string]any{"watched": true}, &viewer, nil), 200)
	continuing := request(t, s, "GET", "/api/playback/continue?type=tv", nil, &viewer, nil)
	status(t, continuing, 200)
	if !strings.Contains(continuing.Body.String(), `"season_number":2`) {
		t.Fatal("continue did not advance across seasons")
	}
	movies := request(t, s, "GET", "/api/media", nil, &viewer, nil)
	status(t, movies, 200)
	if !strings.Contains(movies.Body.String(), `"total":0`) {
		t.Fatal("TV files leaked into movie list")
	}
	for _, route := range []string{"/series", "/series/1"} {
		if page := request(t, s, "GET", route, nil, nil, nil); page.Code != 200 || !strings.Contains(page.Body.String(), `id="root"`) {
			t.Fatal("deep link did not render")
		}
	}
}
