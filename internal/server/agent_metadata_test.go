package server

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"dinosaur/internal/agent"
	"dinosaur/internal/metadata"
)

func TestAgentCanConfirmWithoutProviderIDAndCannotLoseConcurrentUpdates(t *testing.T) {
	s, db, admin, _, _ := agentFixture(t)
	_, token := createTestAgentKey(t, s, &admin, 0, "read_write")
	target := agentRequest(t, s, "GET", "/api/agent/movies/1", nil, token, "")
	status(t, target, 200)
	if target.Header().Get("ETag") != `"0"` || !strings.Contains(target.Body.String(), `"path":"Movie.1999.mkv"`) {
		t.Fatal("missing version or file context")
	}
	status(t, agentRequest(t, s, "GET", "/api/agent/movies/2", nil, token, ""), 404)
	input := agent.Input{Title: "Agent 电影", OriginalTitle: "Original", Overview: "多行\n简介", Year: 1999, ReleaseDate: "1999-06-05", OriginalLanguage: "ja", Language: "zh-CN", Aliases: []string{"Alias", "另一个名称"}, IDs: map[string]string{"imdb": "tt123"}, SourceURL: "https://example.com/source"}
	status(t, agentRequest(t, s, "PUT", "/api/agent/movies/1/metadata", input, token, ""), 428)
	status(t, agentRequest(t, s, "PUT", "/api/agent/movies/1/metadata", input, token, `*`), 400)
	updated := agentRequest(t, s, "PUT", "/api/agent/movies/1/metadata", input, token, `"0"`)
	status(t, updated, 200)
	if updated.Header().Get("ETag") != `"1"` {
		t.Fatal("revision not advanced")
	}
	stale := agentRequest(t, s, "PUT", "/api/agent/movies/1/metadata", agent.Input{Title: "stale"}, token, `"0"`)
	status(t, stale, 409)
	if stale.Header().Get("ETag") != `"1"` {
		t.Fatal("conflict omitted latest version")
	}
	item, err := db.MediaByID(context.Background(), 1)
	if err != nil || item.Title != input.Title || !item.Metadata.Managed() || item.Metadata.TMDBID != 0 || item.OriginalLanguage != "ja" {
		t.Fatal("ID-free confirmed metadata not displayed")
	}
	tasks := agentRequest(t, s, "GET", "/api/agent/tasks", nil, token, "")
	status(t, tasks, 200)
	var page metadata.RecordPage
	if err := json.Unmarshal(tasks.Body.Bytes(), &page); err != nil {
		t.Fatal(err)
	}
	if page.Total != 1 || page.Items[0].SeriesID != 1 {
		t.Fatal("confirmed target still in pending tasks")
	}
	jsonMetadata := agentRequest(t, s, "GET", "/api/agent/movies/1/metadata", nil, token, "")
	status(t, jsonMetadata, 200)
	if !strings.Contains(jsonMetadata.Body.String(), `"imdb":"tt123"`) || !strings.Contains(jsonMetadata.Body.String(), `"aliases":["Alias","另一个名称"]`) {
		t.Fatal("external IDs lost")
	}
	exported := agentRequest(t, s, "GET", "/api/agent/movies/1/metadata.nfo", nil, token, "")
	if exported.Code != 200 {
		t.Fatalf("NFO export status: %d", exported.Code)
	}
	if !strings.HasPrefix(exported.Header().Get("Content-Type"), "application/xml") || !strings.Contains(exported.Body.String(), "<movie>") || !strings.Contains(exported.Body.String(), "Agent 电影") {
		t.Fatal("NFO export invalid")
	}
	updatedNFO := strings.Replace(exported.Body.String(), "Agent 电影", "NFO 电影", 1)
	status(t, agentRaw(t, s, "PUT", "/api/agent/movies/1/metadata", []byte(updatedNFO), "application/xml", token, `"1"`), 200)
	item, err = db.MediaByID(context.Background(), 1)
	if err != nil || item.Title != "NFO 电影" || item.Metadata.Source != "nfo" || item.Metadata.Version != 2 {
		t.Fatal("NFO import did not use shared versioned model")
	}
	status(t, agentRaw(t, s, "PUT", "/api/agent/series/1/metadata", []byte(updatedNFO), "application/xml", token, `"0"`), 400)
	tvNFO := `<?xml version="1.0"?><tvshow><title>Agent 剧集</title><year>2024</year><dinosaur><original_language>ja</original_language></dinosaur></tvshow>`
	status(t, agentRaw(t, s, "PUT", "/api/agent/series/1/metadata", []byte(tvNFO), "application/xml", token, `"0"`), 200)
	show, err := db.SeriesByID(context.Background(), 1)
	if err != nil || show.Title != "Agent 剧集" || show.Metadata.Source != "nfo" || show.Metadata.Version != 1 {
		t.Fatal("TV NFO not saved")
	}
	episode, err := db.MediaByID(context.Background(), 2)
	if err != nil || episode.OriginalLanguage != "ja" {
		t.Fatal("original audio language did not reach episode playback")
	}
	status(t, agentRaw(t, s, "PUT", "/api/agent/movies/1/metadata", []byte(`{"title":"x","unknown":true}`), "application/json", token, `"2"`), 400)
	status(t, agentRaw(t, s, "PUT", "/api/agent/movies/1/metadata", []byte(strings.Repeat("x", 128*1024+1)), "application/json", token, `"2"`), 413)
}
