package server

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"dinosaur/internal/media"
	"dinosaur/internal/metadata"
	"dinosaur/internal/store"
)

func waitServerScrape(t *testing.T, db *store.Store, id int64) metadata.Job {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		jobs, err := db.ScrapeJobs(context.Background())
		if err != nil {
			t.Fatal(err)
		}
		for _, job := range jobs {
			if job.ID == id && job.Status != "queued" && job.Status != "running" {
				return job
			}
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatal("scraping job did not complete")
	return metadata.Job{}
}

func TestScrapingAPIAndCandidatePermissions(t *testing.T) {
	s, db := fixture(t)
	ctx := context.Background()
	setupAdmin(t, s)
	admin := loginAs(t, s, "admin", testPassword)
	lib, err := db.SaveLibrary(ctx, store.Library{Name: "Movies", Type: "movie", RootPath: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.SaveMedia(ctx, media.Item{LibraryID: lib.ID, RelativePath: "Example.2024.mkv", Title: "Example", Year: 2024}, 1); err != nil {
		t.Fatal(err)
	}
	jobsPath := "/api/admin/metadata/jobs"
	candidatesPath := "/api/admin/media/1/metadata/candidates"
	status(t, request(t, s, "POST", jobsPath, metadata.ScrapeRequest{}, &admin, nil), 400)
	if _, err := db.UpdateMetadataSettings(ctx, metadata.SettingsUpdate{Credential: "private-scrape-key"}, false); err != nil {
		t.Fatal(err)
	}
	var calls atomic.Int32
	s.tmdb.HTTP.Transport = metadataTransport(func(r *http.Request) (*http.Response, error) {
		calls.Add(1)
		if r.Header.Get("Authorization") != "Bearer private-scrape-key" {
			t.Error("missing provider credential")
		}
		body := `{"id":123,"title":"中文电影","original_title":"Example","release_date":"2024-01-01","overview":"中文简介"}`
		if r.URL.Path == "/3/search/movie" {
			body = `{"results":[{"id":123,"title":"中文电影","original_title":"Example","release_date":"2024-01-01"}]}`
		}
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(body))}, nil
	})
	hash, err := s.hasher.Hash(ctx, testPassword)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.CreateUser(ctx, "viewer", hash, "user", false); err != nil {
		t.Fatal(err)
	}
	viewer := loginAs(t, s, "viewer", testPassword)
	for _, who := range []*session{nil, &viewer} {
		want := 401
		if who != nil {
			want = 403
		}
		status(t, request(t, s, "GET", jobsPath, nil, who, nil), want)
		status(t, request(t, s, "POST", jobsPath, metadata.ScrapeRequest{}, who, nil), want)
		status(t, request(t, s, "GET", candidatesPath, nil, who, nil), want)
	}
	status(t, request(t, s, "POST", jobsPath, metadata.ScrapeRequest{}, &admin, map[string]string{"X-CSRF-Token": ""}), 403)
	if calls.Load() != 0 {
		t.Fatal("unauthorized request contacted provider")
	}
	status(t, request(t, s, "GET", candidatesPath+"?q=Example&year=2024", nil, &admin, nil), 200)
	status(t, request(t, s, "GET", candidatesPath+"?q=", nil, &admin, nil), 400)
	status(t, request(t, s, "GET", candidatesPath+"?year=bad", nil, &admin, nil), 400)
	for _, bad := range []metadata.ScrapeRequest{{LibraryID: -1}, {TMDBID: 1}, {Refresh: true}} {
		status(t, request(t, s, "POST", jobsPath, bad, &admin, nil), 400)
	}
	status(t, request(t, s, "POST", jobsPath, metadata.ScrapeRequest{MediaID: 999}, &admin, nil), 404)
	started := request(t, s, "POST", jobsPath, metadata.ScrapeRequest{}, &admin, nil)
	status(t, started, 202)
	var result struct {
		Job metadata.Job `json:"job"`
	}
	if err := json.Unmarshal(started.Body.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	if job := waitServerScrape(t, db, result.Job.ID); job.Matched != 1 {
		t.Fatalf("invalid scraping result: %+v", job)
	}
	detail := request(t, s, "GET", "/api/media/1", nil, &viewer, nil)
	status(t, detail, 200)
	if !strings.Contains(detail.Body.String(), "中文电影") || strings.Contains(detail.Body.String(), "private-scrape-key") {
		t.Fatal("localized metadata missing or secret exposed")
	}
	status(t, request(t, s, "GET", "/api/media?q=中文", nil, &viewer, nil), 200)
	status(t, request(t, s, "GET", jobsPath, nil, &admin, nil), 200)
}
