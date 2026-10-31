package server

import (
	"bytes"
	"image"
	"image/png"
	"testing"

	"dinosaur/internal/agent"
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
}
