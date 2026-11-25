package server

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"testing"

	"dinosaur/internal/history"
	"dinosaur/internal/historyassets"
)

func TestHistoryOwnershipCRUDAndNativeAuthentication(t *testing.T) {
	s, db := fixture(t)
	setupAdmin(t, s)
	admin := loginAs(t, s, "admin", testPassword)
	user, _ := db.UserByName(context.Background(), "admin")
	if _, err := db.CreateUser(context.Background(), "viewer", user.PasswordHash, "user", false); err != nil {
		t.Fatal(err)
	}
	viewer := loginAs(t, s, "viewer", testPassword)
	input := history.Input{Title: "Diary", WatchedOn: "2024-04-21", Rating: 5, Review: "私有短评"}
	status(t, request(t, s, "GET", "/api/history", nil, nil, nil), 401)
	status(t, request(t, s, "POST", "/api/history", input, &admin, map[string]string{"X-CSRF-Token": ""}), 403)
	w := request(t, s, "POST", "/api/history", input, &admin, nil)
	status(t, w, 201)
	var created struct {
		Item history.Entry `json:"item"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &created); err != nil {
		t.Fatal(err)
	}
	path := fmt.Sprintf("/api/history/%d", created.Item.ID)
	status(t, request(t, s, "PUT", path, input, &viewer, nil), 404)
	status(t, request(t, s, "DELETE", path, nil, &viewer, nil), 404)
	status(t, request(t, s, "GET", path+"/poster", nil, &viewer, nil), 404)
	input.Review = "Edited"
	status(t, request(t, s, "PUT", path, input, &admin, nil), 200)
	for _, test := range []struct {
		who   session
		count int
	}{{admin, 1}, {viewer, 0}} {
		w := request(t, s, "GET", "/api/history", nil, &test.who, nil)
		status(t, w, 200)
		var page history.Page
		if err := json.Unmarshal(w.Body.Bytes(), &page); err != nil || page.Total != test.count {
			t.Fatalf("%s %v", w.Body.String(), err)
		}
	}
	native := exchangeMobile(t, s, mobileTicket(t, s, admin))
	status(t, request(t, s, "GET", "/api/history", nil, nil, bearer(native.Token)), 200)
	for _, query := range []string{"?page=0", "?page_size=101", "?page=abc", "?page=1000001"} {
		status(t, request(t, s, "GET", "/api/history"+query, nil, &admin, nil), 400)
	}
	input.WatchedOn = "2024-02-30"
	status(t, request(t, s, "POST", "/api/history", input, &admin, nil), 400)
	status(t, request(t, s, "POST", "/api/history", map[string]any{"user_id": 2, "title": "bad"}, &admin, nil), 400)
	status(t, request(t, s, "DELETE", path, nil, &admin, nil), 200)
}

func TestHistoryPostersAreLocalAndOwnerProtected(t *testing.T) {
	s, db := fixture(t)
	setupAdmin(t, s)
	admin := loginAs(t, s, "admin", testPassword)
	data := []byte("RIFF0000WEBPposter-data")
	key, err := historyassets.Put(filepath.Join(s.config.DataDir, "history-posters"), data)
	if err != nil {
		t.Fatal(err)
	}
	entry := history.Entry{Title: "Imported", WatchedOn: "2024-04-21", Source: "douban", SourceID: "123", PosterKey: key}
	id, err := db.CreateHistory(context.Background(), 1, entry)
	if err != nil {
		t.Fatal(err)
	}
	path := fmt.Sprintf("/api/history/%d/poster", id)
	status(t, request(t, s, "GET", path, nil, nil, nil), 401)
	w := request(t, s, "GET", path, nil, &admin, nil)
	if w.Code != 200 || w.Body.String() != string(data) || w.Header().Get("Content-Type") != "image/webp" {
		t.Fatalf("invalid poster: %d %v", w.Code, w.Header())
	}
	if w.Header().Get("Cache-Control") != "private, no-store" {
		t.Fatal("poster cache is public")
	}
}

func TestMovieHistoryRequiresPlaybackEnd(t *testing.T) {
	s, _, admin, _ := playbackFixture(t)
	for _, step := range []struct {
		body  map[string]any
		total int
	}{
		{map[string]any{"position": 950}, 0},
		{map[string]any{"watched": true}, 0},
		{map[string]any{"position": 1000}, 0},
		{map[string]any{"position": 1000, "watched": true}, 1},
		{map[string]any{"position": 1000, "watched": true}, 1},
	} {
		status(t, request(t, s, "PUT", "/api/playback/1/progress", step.body, &admin, nil), 200)
		w := request(t, s, "GET", "/api/history", nil, &admin, nil)
		status(t, w, 200)
		var page history.Page
		if err := json.Unmarshal(w.Body.Bytes(), &page); err != nil || page.Total != step.total {
			t.Fatalf("step %+v: %s %v", step, w.Body.String(), err)
		}
	}
}
