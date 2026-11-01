package server

import (
	"bytes"
	"encoding/json"
	"image"
	"image/png"
	"strings"
	"testing"

	"dinosaur/internal/agent"
	"dinosaur/internal/metadata"
)

func TestAgentArtworkUploadUsesVersionAndScope(t *testing.T) {
	s, _, admin, movieLibrary, tvLibrary := agentFixture(t)
	_, token := createTestAgentKey(t, s, &admin, movieLibrary, "read_write")
	var data bytes.Buffer
	if err := png.Encode(&data, image.NewRGBA(image.Rect(0, 0, 12, 20))); err != nil {
		t.Fatal(err)
	}
	status(t, agentRaw(t, s, "POST", "/api/agent/movies/1/artwork", data.Bytes(), "image/png", token, `"0"`), 409)
	status(t, agentRequest(t, s, "PUT", "/api/agent/movies/1/metadata", agent.Input{Title: "Confirmed"}, token, `"0"`), 200)
	result := agentRaw(t, s, "POST", "/api/agent/movies/1/artwork", data.Bytes(), "image/png", token, `"1"`)
	status(t, result, 200)
	if result.Header().Get("ETag") != `"2"` {
		t.Fatal("artwork did not advance metadata revision")
	}
	fetched := agentRequest(t, s, "GET", "/api/agent/movies/1/artwork", nil, token, "")
	if fetched.Code != 200 {
		t.Fatalf("artwork fetch status: %d", fetched.Code)
	}
	if fetched.Header().Get("Content-Type") != "image/jpeg" {
		t.Fatal("artwork not normalized to JPEG")
	}
	status(t, agentRaw(t, s, "POST", "/api/agent/movies/1/artwork", data.Bytes(), "image/png", token, `"1"`), 409)
	status(t, agentRaw(t, s, "POST", "/api/agent/movies/1/artwork", []byte("not an image"), "image/png", token, `"2"`), 400)
	_, tvOnly := createTestAgentKey(t, s, &admin, tvLibrary, "read_write")
	status(t, agentRaw(t, s, "POST", "/api/agent/movies/1/artwork", data.Bytes(), "image/png", tvOnly, `"2"`), 403)
	status(t, agentRequest(t, s, "GET", "/api/agent/movies/1/artwork", nil, tvOnly, ""), 403)
	status(t, agentRequest(t, s, "PUT", "/api/agent/series/1/metadata", agent.Input{Title: "Series"}, tvOnly, `"0"`), 200)
	status(t, agentRaw(t, s, "POST", "/api/agent/series/1/artwork", data.Bytes(), "image/png", tvOnly, `"1"`), 200)
	page := request(t, s, "GET", "/api/admin/metadata/records", nil, &admin, nil)
	status(t, page, 200)
	var records metadata.RecordPage
	if err := json.Unmarshal(page.Body.Bytes(), &records); err != nil {
		t.Fatal(err)
	}
	if len(records.Items) != 2 {
		t.Fatalf("unexpected record count: %d", len(records.Items))
	}
	for _, item := range records.Items {
		prefix := "/api/media/1/poster?v="
		if item.SeriesID > 0 {
			prefix = "/api/series/1/poster?v="
		}
		if !strings.HasPrefix(item.PosterURL, prefix) {
			t.Fatalf("record lost artwork URL: %+v", item)
		}
		poster := request(t, s, "GET", item.PosterURL, nil, &admin, nil)
		if poster.Code != 200 || poster.Header().Get("Content-Type") != "image/jpeg" {
			t.Fatal("record artwork is not available to the browser")
		}
	}
}
