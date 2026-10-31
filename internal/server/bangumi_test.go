package server

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	"dinosaur/internal/bangumi"
	"dinosaur/internal/media"
	"dinosaur/internal/metadata"
	"dinosaur/internal/store"
	"dinosaur/internal/television"
)

func TestBangumiAPISourceSelectionAndPermissions(t *testing.T) {
	for _, kind := range []string{"movie", "tv"} {
		t.Run(kind, func(t *testing.T) {
			s, db := fixture(t)
			setupAdmin(t, s)
			admin := loginAs(t, s, "admin", testPassword)
			ctx := context.Background()
			lib, err := db.SaveLibrary(ctx, store.Library{Name: "Library", Type: kind, RootPath: t.TempDir()})
			if err != nil {
				t.Fatal(err)
			}
			if kind == "movie" {
				err = db.SaveMedia(ctx, media.Item{LibraryID: lib.ID, RelativePath: "Example.2024.mkv", Title: "Example", Year: 2024}, 1)
			} else {
				path := "Example (2024)/Example.S01E01.mkv"
				name, _ := television.ParseEpisode(path)
				err = db.SaveEpisode(ctx, media.Item{LibraryID: lib.ID, RelativePath: path, Title: "Example"}, name, 1)
			}
			if err != nil {
				t.Fatal(err)
			}
			client := bangumi.New(kind)
			client.HTTP.Transport = metadataTransport(func(r *http.Request) (*http.Response, error) {
				if r.Header.Get("Authorization") != "" {
					t.Error("TMDB credential leaked")
				}
				platform := "剧场版"
				if kind == "tv" {
					platform = "TV"
				}
				body := `{"id":123,"type":2,"name":"Example","name_cn":"中文标题","date":"2024-01-01","platform":"` + platform + `"}`
				if r.Method == http.MethodPost {
					body = `{"data":[{"id":123,"type":2}]}`
				}
				return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(body))}, nil
			})
			jobs, candidates, detail := "/api/admin/metadata/jobs", "/api/admin/media/1/metadata/candidates", "/api/media/1"
			body := map[string]any{"media_id": 1, "bangumi_id": 123}
			if kind == "tv" {
				s.tvScraper.WithFallback(client)
				jobs, candidates, detail = "/api/admin/tv/metadata/jobs", "/api/admin/series/1/metadata/candidates", "/api/series/1"
				body = map[string]any{"series_id": 1, "bangumi_id": 123}
			} else {
				s.scraper.WithFallback(client)
			}
			status(t, request(t, s, "GET", candidates, nil, nil, nil), 401)
			status(t, request(t, s, "POST", jobs, body, &admin, map[string]string{"X-CSRF-Token": ""}), 403)
			found := request(t, s, "GET", candidates, nil, &admin, nil)
			status(t, found, 200)
			if !strings.Contains(found.Body.String(), `"provider":"bangumi"`) {
				t.Fatal("candidate provenance missing")
			}
			started := request(t, s, "POST", jobs, body, &admin, nil)
			status(t, started, 202)
			var result struct {
				Job metadata.Job `json:"job"`
			}
			if err := json.Unmarshal(started.Body.Bytes(), &result); err != nil {
				t.Fatal(err)
			}
			if job := waitServerScrape(t, db, result.Job.ID); job.Matched != 1 {
				t.Fatalf("job: %+v", job)
			}
			response := request(t, s, "GET", detail, nil, &admin, nil)
			status(t, response, 200)
			if !strings.Contains(response.Body.String(), `"bangumi_id":123`) || !strings.Contains(response.Body.String(), `"tmdb_id":0`) {
				t.Fatal("metadata IDs conflated")
			}
		})
	}
}
