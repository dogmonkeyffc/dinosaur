package server

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"dinosaur/internal/media"
	"dinosaur/internal/playback"
	"dinosaur/internal/store"
)

func TestLibraryAccessAppliesToCatalogAndDirectRequests(t *testing.T) {
	s, db, admin, viewer, tv, seriesIDs := workHTTPFixture(t)
	ctx := context.Background()
	user, err := db.UserByName(ctx, "work-viewer")
	if err != nil {
		t.Fatal(err)
	}
	phone := exchangeMobile(t, s, mobileTicket(t, s, viewer))
	work, err := db.SaveWork(ctx, 0, "Limited work", seriesIDs)
	if err != nil {
		t.Fatal(err)
	}
	root := t.TempDir()
	movie, err := db.SaveLibrary(ctx, store.Library{Name: "Limited movies", Type: "movie", RootPath: root})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "Private.mp4"), []byte("video"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := db.SaveMedia(ctx, media.Item{LibraryID: movie.ID, Title: "Private", RelativePath: "Private.mp4",
		Filename: "Private.mp4", VideoCodec: "h264", Duration: 100}, 1); err != nil {
		t.Fatal(err)
	}
	var mediaID int64
	if err := db.DB.QueryRow("SELECT id FROM media_items WHERE library_id=?", movie.ID).Scan(&mediaID); err != nil {
		t.Fatal(err)
	}
	if err := db.SavePlaybackProgress(ctx, user.ID, playback.Progress{MediaID: mediaID, Position: 20, UpdatedAt: 1}); err != nil {
		t.Fatal(err)
	}
	moviePath := fmt.Sprintf("/api/media/%d", mediaID)
	workPath := fmt.Sprintf("/api/works/%d", work.ID)
	seriesPath := fmt.Sprintf("/api/series/%d", seriesIDs[0])
	playPath := fmt.Sprintf("/api/playback/%d", mediaID)
	for _, id := range []int64{tv, movie.ID} {
		path := fmt.Sprintf("/api/admin/libraries/%d/access", id)
		status(t, request(t, s, "GET", path, nil, &viewer, nil), 403)
		status(t, request(t, s, "PUT", path, store.LibraryAccess{Restricted: true}, &viewer, nil), 403)
		status(t, request(t, s, "PUT", path, store.LibraryAccess{Restricted: true}, &admin, nil), 204)
		status(t, request(t, s, "PUT", path, store.LibraryAccess{UserIDs: []int64{999999}}, &admin, nil), 400)
	}
	for _, path := range []string{moviePath, moviePath + "/poster", seriesPath, seriesPath + "/poster", workPath, playPath, playPath + "/stream"} {
		status(t, request(t, s, "GET", path, nil, nil, bearer(phone.Token)), 404)
	}
	status(t, request(t, s, "PUT", playPath+"/progress", map[string]int{"position": 21}, &viewer, nil), 404)
	for _, tc := range []struct{ path, field string }{
		{"/api/libraries", "libraries"}, {"/api/media", "items"}, {"/api/series", "items"},
		{"/api/series?raw=1", "items"}, {"/api/works", "items"}, {"/api/playback/continue", "items"},
	} {
		response := request(t, s, "GET", tc.path, nil, &viewer, nil)
		status(t, response, 200)
		var body map[string]json.RawMessage
		if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil || string(body[tc.field]) != "[]" {
			t.Fatalf("%s exposed restricted data: %s", tc.path, response.Body)
		}
	}
	for _, path := range []string{moviePath, seriesPath, workPath} {
		status(t, request(t, s, "GET", path, nil, &admin, nil), 200)
	}
	accessPath := fmt.Sprintf("/api/admin/libraries/%d/access", movie.ID)
	status(t, request(t, s, "PUT", accessPath, store.LibraryAccess{Restricted: true, UserIDs: []int64{user.ID}}, &admin, nil), 204)
	status(t, request(t, s, "GET", moviePath, nil, nil, bearer(phone.Token)), 200)
	status(t, request(t, s, "GET", moviePath, nil, &viewer, nil), 200)
	status(t, request(t, s, "PUT", accessPath, store.LibraryAccess{Restricted: true}, &admin, nil), 204)
	status(t, request(t, s, "GET", moviePath, nil, &viewer, nil), 404)
}
