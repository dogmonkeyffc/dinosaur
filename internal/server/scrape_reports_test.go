package server

import (
	"context"
	"testing"

	"dinosaur/internal/media"
	"dinosaur/internal/metadata"
	"dinosaur/internal/store"
)

func TestScrapeReportsRequireAdminAndValidateFilters(t *testing.T) {
	s, db := fixture(t)
	setupAdmin(t, s)
	admin := loginAs(t, s, "admin", testPassword)
	hash, err := s.hasher.Hash(context.Background(), testPassword)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.CreateUser(context.Background(), "viewer", hash, "user", false); err != nil {
		t.Fatal(err)
	}
	viewer := loginAs(t, s, "viewer", testPassword)
	lib, err := db.SaveLibrary(context.Background(), store.Library{Name: "Movies", Type: "movie", RootPath: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.SaveMedia(context.Background(), media.Item{LibraryID: lib.ID, Title: "Movie", RelativePath: "Movie.mp4"}, 1); err != nil {
		t.Fatal(err)
	}
	job, err := db.NewScrapeJob(context.Background(), metadata.ScrapeRequest{})
	if err != nil || job.ID != 1 {
		t.Fatal("missing fixture job")
	}
	for _, path := range []string{"/api/admin/metadata/records", "/api/admin/metadata/jobs", "/api/admin/metadata/jobs/1"} {
		status(t, request(t, s, "GET", path, nil, nil, nil), 401)
		status(t, request(t, s, "GET", path, nil, &viewer, nil), 403)
		status(t, request(t, s, "GET", path, nil, &admin, nil), 200)
		for _, query := range []string{"?page=0", "?page_size=101", "?status=unknown", "?status=matched%20unmatched", "?library_id=-1"} {
			status(t, request(t, s, "GET", path+query, nil, &admin, nil), 400)
		}
	}
	status(t, request(t, s, "GET", "/api/admin/metadata/jobs/999", nil, &admin, nil), 404)
	status(t, request(t, s, "GET", "/api/admin/metadata/jobs/invalid", nil, &admin, nil), 400)
}
