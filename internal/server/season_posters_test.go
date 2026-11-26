package server

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"image"
	"image/jpeg"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"dinosaur/internal/media"
	"dinosaur/internal/store"
	"dinosaur/internal/television"
)

func TestSeasonPostersBackfillAndFallback(t *testing.T) {
	s, db := fixture(t)
	setupAdmin(t, s)
	admin := loginAs(t, s, "admin", testPassword)
	ctx := context.Background()
	lib, err := db.SaveLibrary(ctx, store.Library{Name: "Series", Type: "tv", RootPath: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	for series := 1; series <= 2; series++ {
		for season := 1; season <= 2; season++ {
			file := filepath.Join(fmt.Sprintf("Example %d", series), fmt.Sprintf("Example.S%02dE01.mp4", season))
			episode := television.EpisodeName{Key: fmt.Sprintf("example-%d", series), Series: fmt.Sprintf("Example %d", series), Season: season, Episode: 1, End: 1}
			err = db.SaveEpisode(ctx, media.Item{LibraryID: lib.ID, RelativePath: file, Title: "Episode"}, episode, 1)
			if err != nil {
				t.Fatal(err)
			}
		}
	}
	if _, err = db.DB.ExecContext(ctx, `UPDATE tv_series SET metadata_json=json_object('tmdb_id',id+100,'status','matched')`); err != nil {
		t.Fatal(err)
	}
	if _, err = db.DB.ExecContext(ctx, `UPDATE metadata_settings SET credential='test-key',credential_type='api_key'`); err != nil {
		t.Fatal(err)
	}
	var poster bytes.Buffer
	if err := jpeg.Encode(&poster, image.NewRGBA(image.Rect(0, 12, 8, 24)), nil); err != nil {
		t.Fatal(err)
	}
	s.tmdb.HTTP.Transport = metadataTransport(func(r *http.Request) (*http.Response, error) {
		var body io.Reader
		switch {
		case strings.HasPrefix(r.URL.Path, "/3/tv/"):
			body = strings.NewReader(fmt.Sprintf(`{"id":%s,"seasons":[{"season_number":1,"poster_path":"/Artwork.jpg"}]}`,
				strings.TrimPrefix(r.URL.Path, "/3/tv/")))
		case r.URL.Path == "/t/p/w500/Artwork.jpg":
			body = bytes.NewReader(poster.Bytes())
		default:
			return nil, fmt.Errorf("unexpected request path %s", r.URL.Path)
		}
		return &http.Response{StatusCode: 200, Body: io.NopCloser(body)}, nil
	})
	targets, err := db.PendingSeasonPosters(ctx, 8)
	if err != nil || len(targets) != 2 {
		t.Fatalf("pending targets: %v %v", targets, err)
	}
	if err := s.RecoverInterruptedJobs(ctx); err != nil {
		t.Fatal(err)
	}
	deadline := time.Now().Add(3 * time.Second)
	for len(targets) > 0 && time.Now().Before(deadline) {
		time.Sleep(10 * time.Millisecond)
		targets, err = db.PendingSeasonPosters(ctx, 8)
		if err != nil {
			t.Fatal(err)
		}
	}
	if len(targets) != 0 {
		t.Fatalf("posters were not processed: %v", targets)
	}
	for id := 1; id <= 2; id++ {
		base := fmt.Sprintf("/api/series/%d", id)
		w := request(t, s, "GET", base, nil, &admin, nil)
		status(t, w, 200)
		var detail television.Detail
		if err := json.Unmarshal(w.Body.Bytes(), &detail); err != nil {
			t.Fatal(err)
		}
		if len(detail.Seasons) != 2 || !strings.HasPrefix(detail.Seasons[0].PosterURL, base+"/seasons/1/poster") || detail.Seasons[1].PosterURL != "" {
			t.Fatalf("unexpected season covers: %+v", detail.Seasons)
		}
		w = request(t, s, "GET", detail.Seasons[0].PosterURL, nil, &admin, nil)
		if w.Code != 200 || !bytes.Equal(w.Body.Bytes(), poster.Bytes()) {
			t.Fatalf("wrong cached poster: %d", w.Code)
		}
		key := strings.Split(detail.Seasons[0].PosterURL, "?v=")[1]
		if _, err := db.DB.ExecContext(ctx, `UPDATE tv_series SET poster_key=? WHERE id=?`, key, id); err != nil {
			t.Fatal(err)
		}
		w = request(t, s, "GET", base+"/seasons/2/poster", nil, &admin, nil)
		if w.Code != 200 || !bytes.Equal(w.Body.Bytes(), poster.Bytes()) {
			t.Fatalf("wrong fallback cover: %d", w.Code)
		}
		status(t, request(t, s, "GET", base+"/seasons/1/poster", nil, nil, nil), 401)
	}
}
