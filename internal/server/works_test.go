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

func workHTTPFixture(t *testing.T) (*Server, *store.Store, session, session, int64, []int64) {
	t.Helper()
	s, db := fixture(t)
	setupAdmin(t, s)
	admin := loginAs(t, s, "admin", testPassword)
	ctx := context.Background()
	user, err := db.UserByName(ctx, "admin")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.CreateUser(ctx, "work-viewer", user.PasswordHash, "user", false); err != nil {
		t.Fatal(err)
	}
	viewer := loginAs(t, s, "work-viewer", testPassword)
	lib, err := db.SaveLibrary(ctx, store.Library{Name: "TV", Type: "tv", RootPath: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	var ids []int64
	for _, path := range []string{"Alpha/S01E01.mp4", "Other/Beta/S02E01.mp4", "Gamma/S01E01.mp4"} {
		name, ok := television.ParseEpisode(path)
		if !ok {
			t.Fatal(path)
		}
		if err := db.SaveEpisode(ctx, media.Item{LibraryID: lib.ID, RelativePath: path, Title: name.Title}, name, 1); err != nil {
			t.Fatal(err)
		}
		var id int64
		if err := db.DB.QueryRow(`SELECT e.series_id FROM media_items m JOIN tv_episodes e ON e.media_id=m.id WHERE m.library_id=? AND m.relative_path=?`, lib.ID, path).Scan(&id); err != nil {
			t.Fatal(err)
		}
		ids = append(ids, id)
	}
	return s, db, admin, viewer, lib.ID, ids
}

func TestWorkHTTPCreateReplaceBrowseAndDissolve(t *testing.T) {
	s, db, admin, viewer, lib, ids := workHTTPFixture(t)
	if _, err := db.DB.Exec(`UPDATE tv_series SET poster_key='aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa' WHERE id=?`, ids[1]); err != nil {
		t.Fatal(err)
	}
	body := map[string]any{"title": "Anthology", "series_ids": []int64{ids[1], ids[0]}}
	created := request(t, s, "POST", "/api/admin/works", body, &admin, nil)
	status(t, created, 201)
	var result struct {
		Work television.Work `json:"work"`
	}
	if err := json.Unmarshal(created.Body.Bytes(), &result); err != nil || result.Work.ID == 0 || result.Work.Title != "Anthology" || result.Work.LibraryID != lib || result.Work.MemberCount != 2 || result.Work.PosterURL == "" {
		t.Fatalf("create=%s error=%v", created.Body.String(), err)
	}
	path := fmt.Sprintf("/api/works/%d", result.Work.ID)
	adminPath := fmt.Sprintf("/api/admin/works/%d", result.Work.ID)
	listed := request(t, s, "GET", fmt.Sprintf("/api/works?library_id=%d", lib), nil, &viewer, nil)
	status(t, listed, 200)
	var listing struct {
		Items []television.Work `json:"items"`
	}
	if err := json.Unmarshal(listed.Body.Bytes(), &listing); err != nil || len(listing.Items) != 1 || listing.Items[0].ID != result.Work.ID {
		t.Fatalf("list=%s error=%v", listed.Body.String(), err)
	}
	for _, who := range []*session{&viewer, &admin} {
		response := request(t, s, "GET", path, nil, who, nil)
		status(t, response, 200)
		var detail television.WorkDetail
		if err := json.Unmarshal(response.Body.Bytes(), &detail); err != nil || len(detail.Members) != 2 || detail.Members[0].ID != ids[1] || detail.Members[1].ID != ids[0] {
			t.Fatalf("detail=%s error=%v", response.Body.String(), err)
		}
		var fields map[string]json.RawMessage
		if err := json.Unmarshal(response.Body.Bytes(), &fields); err != nil {
			t.Fatal(err)
		}
		_, hasAll := fields["all_members"]
		if hasAll != (who == &admin) {
			t.Fatalf("all_members authorization: %s", response.Body.String())
		}
	}
	for _, raw := range []bool{false, true} {
		query := "/api/series"
		want := 2
		if raw {
			query += "?raw=1"
			want = 3
		}
		response := request(t, s, "GET", query, nil, &viewer, nil)
		status(t, response, 200)
		var page television.Page
		if err := json.Unmarshal(response.Body.Bytes(), &page); err != nil || page.Total != want {
			t.Fatalf("raw=%v response=%s error=%v", raw, response.Body.String(), err)
		}
	}
	if _, err := db.DB.Exec(`UPDATE media_items SET missing=1 WHERE id IN (SELECT media_id FROM tv_episodes WHERE series_id=?)`, ids[1]); err != nil {
		t.Fatal(err)
	}
	response := request(t, s, "GET", path, nil, &admin, nil)
	status(t, response, 200)
	var detail television.WorkDetail
	if err := json.Unmarshal(response.Body.Bytes(), &detail); err != nil || len(detail.Members) != 1 || len(detail.AllMembers) != 2 || detail.AllMembers[0].ID != ids[1] {
		t.Fatalf("missing member absent from admin editing payload=%s error=%v", response.Body.String(), err)
	}
	body["title"], body["series_ids"] = "Replacement", []int64{ids[2]}
	status(t, request(t, s, "PUT", adminPath, body, &admin, nil), 200)
	response = request(t, s, "GET", path, nil, &viewer, nil)
	status(t, response, 200)
	if err := json.Unmarshal(response.Body.Bytes(), &detail); err != nil || detail.Work.Title != "Replacement" || len(detail.Members) != 1 || detail.Members[0].ID != ids[2] {
		t.Fatalf("replacement=%s error=%v", response.Body.String(), err)
	}
	deleted := request(t, s, "DELETE", adminPath, nil, &admin, nil)
	if deleted.Code != 204 || deleted.Body.Len() != 0 {
		t.Fatalf("delete=%d %s", deleted.Code, deleted.Body.String())
	}
	status(t, request(t, s, "GET", path, nil, &viewer, nil), 404)
	page, err := db.SeriesPage(context.Background(), media.Filter{Page: 1, PageSize: 24}, 0)
	if err != nil || page.Total != 2 {
		t.Fatalf("delete removed series rather than dissolving group: %+v %v", page, err)
	}
}
