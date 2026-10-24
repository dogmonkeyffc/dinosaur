package server

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPlaybackSubtitles(t *testing.T) {
	s, _, admin, root := playbackFixture(t)
	if err := os.WriteFile(filepath.Join(root, "Movie.en.srt"), []byte("1\n00:00:01,000 --> 00:00:03,000\nHello\n"), 0600); err != nil {
		t.Fatal(err)
	}
	outside := filepath.Join(t.TempDir(), "outside.vtt")
	if err := os.WriteFile(outside, []byte("WEBVTT\n\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, filepath.Join(root, "Movie.secret.vtt")); err != nil {
		t.Fatal(err)
	}
	w := request(t, s, "GET", "/api/playback/1", nil, &admin, nil)
	status(t, w, 200)
	if !strings.Contains(w.Body.String(), `"label":"en"`) || strings.Contains(w.Body.String(), "secret") {
		t.Fatal("incorrect subtitle discovery")
	}
	path := "/api/playback/1/subtitles/0"
	status(t, request(t, s, "GET", path, nil, nil, nil), 401)
	w = request(t, s, "GET", path, nil, &admin, nil)
	if w.Code != 200 || !strings.HasPrefix(w.Header().Get("Content-Type"), "text/vtt") || !strings.Contains(w.Body.String(), "00:00:01.000") {
		t.Fatalf("subtitle=%d %s", w.Code, w.Body.String())
	}
	status(t, request(t, s, "GET", "/api/playback/1/subtitles/999", nil, &admin, nil), 404)
}
