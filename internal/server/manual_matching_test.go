package server

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"

	"dinosaur/internal/media"
	"dinosaur/internal/metadata"
	"dinosaur/internal/store"
)

func TestManualCandidateAcceptedWhileBatchIsRunning(t *testing.T) {
	s, db := fixture(t)
	setupAdmin(t, s)
	admin := loginAs(t, s, "admin", testPassword)
	ctx := context.Background()
	lib, err := db.SaveLibrary(ctx, store.Library{Name: "Movies", Type: "movie", RootPath: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.SaveMedia(ctx, media.Item{LibraryID: lib.ID, RelativePath: "Example.2024.mkv", Title: "Example", Year: 2024}, 1); err != nil {
		t.Fatal(err)
	}
	if _, err := db.UpdateMetadataSettings(ctx, metadata.SettingsUpdate{Credential: "test-key"}, false); err != nil {
		t.Fatal(err)
	}
	started, release := make(chan struct{}), make(chan struct{})
	var once sync.Once
	defer once.Do(func() { close(release) })
	s.tmdb.HTTP.Transport = metadataTransport(func(r *http.Request) (*http.Response, error) {
		var body string
		if r.URL.Path == "/3/search/movie" {
			close(started)
			select {
			case <-release:
			case <-r.Context().Done():
				return nil, r.Context().Err()
			}
			body = `{"results":[{"id":10,"title":"Example","release_date":"2024-01-01"}]}`
		} else {
			id := 10
			if strings.HasSuffix(r.URL.Path, "/20") {
				id = 20
			}
			data, _ := json.Marshal(map[string]any{"id": id, "title": "Selected", "release_date": "2024-01-01"})
			body = string(data)
		}
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(body))}, nil
	})
	status(t, request(t, s, "POST", "/api/admin/metadata/jobs", metadata.ScrapeRequest{}, &admin, nil), 202)
	select {
	case <-started:
	case <-time.After(time.Second):
		t.Fatal("batch did not start")
	}
	result := request(t, s, "POST", "/api/admin/metadata/jobs", metadata.ScrapeRequest{MediaID: 1, TMDBID: 20}, &admin, nil)
	status(t, result, 202)
	var selected struct {
		Job metadata.Job `json:"job"`
	}
	if err := json.Unmarshal(result.Body.Bytes(), &selected); err != nil {
		t.Fatal(err)
	}
	if selected.Job.Status != "queued" {
		t.Fatal("manual job was not queued")
	}
	once.Do(func() { close(release) })
	if job := waitServerScrape(t, db, selected.Job.ID); job.Matched != 1 {
		t.Fatalf("manual job: %+v", job)
	}
	item, err := db.MediaByID(ctx, 1)
	if err != nil || item.Metadata.TMDBID != 20 {
		t.Fatalf("manual choice lost: %+v %v", item, err)
	}
}
