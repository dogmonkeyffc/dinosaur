package server

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"dinosaur/internal/media"
	"dinosaur/internal/store"
	"dinosaur/internal/television"
)

func TestWorkHTTPPermissionsCSRFAndBadInput(t *testing.T) {
	s, db, admin, viewer, _, ids := workHTTPFixture(t)
	body := map[string]any{"title": "Group", "series_ids": ids[:1]}
	work, err := db.SaveWork(context.Background(), 0, "Existing", ids[:1])
	if err != nil {
		t.Fatal(err)
	}
	path := fmt.Sprintf("/api/admin/works/%d", work.ID)
	for _, endpoint := range []string{"/api/works", fmt.Sprintf("/api/works/%d", work.ID)} {
		status(t, request(t, s, "GET", endpoint, nil, nil, nil), 401)
	}
	for _, op := range []struct{ method, path string }{
		{"POST", "/api/admin/works"}, {"PUT", path}, {"DELETE", path},
	} {
		t.Run(op.method, func(t *testing.T) {
			status(t, request(t, s, op.method, op.path, body, nil, nil), 401)
			status(t, request(t, s, op.method, op.path, body, &viewer, nil), 403)
			status(t, request(t, s, op.method, op.path, body, &admin, map[string]string{"X-CSRF-Token": ""}), 403)
			status(t, request(t, s, op.method, op.path, body, &admin, map[string]string{"X-CSRF-Token": "wrong"}), 403)
		})
	}
	if _, err := db.WorkByID(context.Background(), work.ID, 0); err != nil {
		t.Fatalf("unauthorized mutation changed work: %v", err)
	}
	for _, endpoint := range []string{"/api/works?library_id=-1", "/api/works?library_id=abc", "/api/works/not-a-number", "/api/works/0"} {
		status(t, request(t, s, "GET", endpoint, nil, &viewer, nil), 400)
	}
	status(t, request(t, s, "GET", "/api/works/999999", nil, &viewer, nil), 404)
	status(t, request(t, s, "PUT", "/api/admin/works/999999", map[string]any{"title": "Unknown", "series_ids": ids[1:]}, &admin, nil), 404)
	status(t, request(t, s, "DELETE", "/api/admin/works/999999", nil, &admin, nil), 404)
	for _, payload := range []any{
		map[string]any{},
		map[string]any{"title": " \t", "series_ids": ids[1:]},
		map[string]any{"title": "Bad", "series_ids": []int64{}},
		map[string]any{"title": "Bad", "series_ids": []int64{0}},
		map[string]any{"title": "Bad", "series_ids": []int64{-1}},
		map[string]any{"title": "Bad", "series_ids": []int64{ids[1], ids[1]}},
		map[string]any{"title": "Bad", "series_ids": "not-an-array"},
		map[string]any{"title": 123, "series_ids": ids[1:]},
	} {
		status(t, request(t, s, "POST", "/api/admin/works", payload, &admin, nil), 400)
	}
	ctx := context.Background()
	other, err := db.SaveLibrary(ctx, store.Library{Name: "Other", Type: "tv", RootPath: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	name, _ := television.ParseEpisode("Foreign/S01E01.mp4")
	if err := db.SaveEpisode(ctx, media.Item{LibraryID: other.ID, RelativePath: "Foreign/S01E01.mp4", Title: name.Title}, name, 1); err != nil {
		t.Fatal(err)
	}
	var foreign int64
	if err := db.DB.QueryRow(`SELECT id FROM tv_series WHERE library_id=?`, other.ID).Scan(&foreign); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		ids  []int64
		code string
		http int
	}{
		{[]int64{ids[1], foreign}, "cross_library", 400},
		{[]int64{ids[1], 999999}, "series_not_found", 400},
		{[]int64{ids[1], ids[0]}, "work_conflict", 409},
		{[]int64{ids[1], ids[1]}, "invalid_work", 400},
	} {
		response := request(t, s, "POST", "/api/admin/works", map[string]any{"title": "Bad", "series_ids": tc.ids}, &admin, nil)
		status(t, response, tc.http)
		var result struct {
			Error struct {
				Code    string `json:"code"`
				Message string `json:"message"`
			} `json:"error"`
		}
		if err := json.Unmarshal(response.Body.Bytes(), &result); err != nil || result.Error.Code != tc.code || result.Error.Message == "" {
			t.Fatalf("error contract=%s want=%s error=%v", response.Body.String(), tc.code, err)
		}
	}
}
