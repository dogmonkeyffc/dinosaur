package server

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"dinosaur/internal/media"
	"dinosaur/internal/playback"
	"dinosaur/internal/store"
	"dinosaur/internal/television"
)

func TestHomeResumeMergesTypesAndFiltersVisibleLibraries(t *testing.T) {
	s, db := fixture(t)
	setupAdmin(t, s)
	admin := loginAs(t, s, "admin", testPassword)
	viewer := createManagedAccount(t, s, db, "home-viewer", admin)
	who := loginAs(t, s, viewer.Username, testPassword)
	ctx := context.Background()

	library := func(name, kind string) int64 {
		t.Helper()
		item, err := db.SaveLibrary(ctx, store.Library{Name: name, Type: kind, RootPath: t.TempDir()})
		if err != nil {
			t.Fatal(err)
		}
		return item.ID
	}
	indexed := func(lib int64, path string, series bool) int64 {
		t.Helper()
		item := media.Item{LibraryID: lib, RelativePath: path, Title: path, Duration: 100}
		var err error
		if series {
			name, ok := television.ParseEpisode(path)
			if !ok {
				t.Fatal(path)
			}
			err = db.SaveEpisode(ctx, item, name, 1)
		} else {
			err = db.SaveMedia(ctx, item, 1)
		}
		if err != nil {
			t.Fatal(err)
		}
		var id int64
		if err := db.DB.QueryRow("SELECT id FROM media_items WHERE library_id=? AND relative_path=?", lib, path).Scan(&id); err != nil {
			t.Fatal(err)
		}
		return id
	}
	movies := library("Movies", "movie")
	videos := library("Videos", "video")
	series := library("Series", "tv")
	hidden := library("Hidden", "movie")
	restricted := library("Restricted", "movie")
	silent := library("No recording", "video")
	if err := db.SaveLibraryAccess(ctx, restricted, store.LibraryAccess{Restricted: true}); err != nil {
		t.Fatal(err)
	}
	save := func(id, updated int64, watched bool) {
		t.Helper()
		if err := db.SavePlaybackProgress(ctx, viewer.ID, playback.Progress{MediaID: id, Position: 20, Watched: watched}); err != nil {
			t.Fatal(err)
		}
		if _, err := db.DB.Exec("UPDATE playback_progress SET updated_at=? WHERE user_id=? AND media_id=?", updated, viewer.ID, id); err != nil {
			t.Fatal(err)
		}
	}
	first := indexed(series, "Show/S01E01.mp4", true)
	next := indexed(series, "Show/S01E02.mp4", true)
	movie := indexed(movies, "Movie.mp4", false)
	clip := indexed(videos, "Clip.mp4", false)
	save(movie, 80, false)
	save(first, 90, true)
	save(clip, 60, false)
	save(indexed(restricted, "Restricted.mp4", false), 200, false)
	save(indexed(silent, "Silent.mp4", false), 160, false)
	if err := db.SetHistoryLibrarySetting(ctx, viewer.ID, silent, false); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 14; i++ {
		save(indexed(hidden, fmt.Sprintf("Hidden-%d.mp4", i), false), int64(100+i), false)
	}
	read := func(query string) []media.Item {
		t.Helper()
		w := request(t, s, "GET", "/api/playback/continue?"+query, nil, &who, nil)
		status(t, w, 200)
		var response struct{ Items []media.Item }
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatal(err)
		}
		return response.Items
	}
	ids := fmt.Sprintf("%d,%d,%d,%d,%d", movies, videos, series, restricted, silent)
	items := read("type=all&library_ids=" + ids)
	if len(items) != 3 || items[0].ID != next || items[0].Position != 0 || items[0].LastPlayedAt != 90 ||
		items[1].ID != movie || items[2].ID != clip {
		t.Fatalf("unexpected resume order: %+v", items)
	}
	if len(read("type=all&library_ids=")) != 0 {
		t.Fatal("explicit empty scope returned items")
	}
	for _, item := range read("type=all") {
		if item.LibraryID == restricted || item.LibraryID == silent {
			t.Fatal("resume exposed unavailable libraries")
		}
	}
	status(t, request(t, s, "GET", "/api/playback/continue?type=all&library_ids=0", nil, &who, nil), 400)
	status(t, request(t, s, "GET", "/api/playback/continue?type=all", nil, nil, nil), 401)
	legacy := read("type=movie")
	if len(legacy) == 0 || legacy[0].Type != "movie" || !strings.Contains(legacy[0].Filename, "Hidden-") {
		t.Fatal("typed scope changed")
	}
}
