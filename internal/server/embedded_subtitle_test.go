package server

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"dinosaur/internal/media"
	"dinosaur/internal/playback"
	"dinosaur/internal/streaming"
)

func TestEmbeddedSubtitleDiscoveryAndBitmapPlanPermissions(t *testing.T) {
	s, db, admin, _ := playbackFixture(t)
	item, err := db.MediaByID(context.Background(), 1)
	if err != nil {
		t.Fatal(err)
	}
	item.Tracks = append(item.Tracks, media.Track{Index: 2, Type: "subtitle", Codec: "subrip", Language: "chi"}, media.Track{Index: 3, Type: "subtitle", Codec: "hdmv_pgs_subtitle", Language: "chi"})
	if err := db.SaveMedia(context.Background(), item, 1); err != nil {
		t.Fatal(err)
	}
	response := request(t, s, "GET", "/api/playback/1?mode=original&subtitle=-4", nil, &admin, nil)
	status(t, response, 200)
	var info struct {
		Plan      playback.Plan        `json:"plan"`
		URL       string               `json:"stream_url"`
		Subtitles []streaming.Subtitle `json:"subtitles"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &info); err != nil {
		t.Fatal(err)
	}
	if info.Plan.Mode != "transcode" || info.Plan.Subtitle == nil || *info.Plan.Subtitle != 3 || !strings.Contains(info.URL, "subtitle=-4") || len(info.Subtitles) != 2 {
		t.Fatalf("bitmap request was not preserved: %+v", info)
	}

	status(t, request(t, s, "GET", info.URL, nil, nil, nil), 401)
	for _, query := range []string{"subtitle=0", "subtitle=-2", "subtitle=-3", "subtitle=-9999", "subtitle=nope"} {
		status(t, request(t, s, "GET", "/api/playback/1?"+query, nil, &admin, nil), 400)
	}
	status(t, request(t, s, "GET", "/api/playback/1/subtitles/-4", nil, &admin, nil), 422)
	status(t, request(t, s, "GET", "/api/playback/1/subtitles/-3", nil, nil, nil), 401)
	for _, start := range []string{"NaN", "-1", "Inf", "86401"} {
		status(t, request(t, s, "GET", "/api/playback/1/subtitles/-3?start="+start, nil, &admin, nil), 400)
	}
	off := request(t, s, "GET", "/api/playback/1?mode=original", nil, &admin, nil)
	status(t, off, 200)
	if strings.Contains(off.Body.String(), `"subtitle":3`) {
		t.Fatal("bitmap subtitle remained after being disabled")
	}
}

func TestEmbeddedSubtitleHTTPReturnsAuthenticatedVTT(t *testing.T) {
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		t.Skip("FFmpeg not installed")
	}
	s, db, admin, root := playbackFixture(t)
	captions := filepath.Join(root, "captions.srt")
	if err := os.WriteFile(captions, []byte("1\n00:00:01,000 --> 00:00:03,000\n字幕验证\n"), 0600); err != nil {
		t.Fatal(err)
	}
	args := []string{"-v", "error", "-y", "-f", "lavfi", "-i", "color=s=64x48:d=5", "-i", captions, "-c:v", "libx264", "-preset", "ultrafast", "-threads", "1", "-c:s", "mov_text", filepath.Join(root, "Movie.mp4")}
	if out, err := exec.Command("ffmpeg", args...).CombinedOutput(); err != nil {
		t.Fatalf("fixture: %v %s", err, out)
	}
	item, err := db.MediaByID(context.Background(), 1)
	if err != nil {
		t.Fatal(err)
	}
	item.Tracks = []media.Track{{Index: 1, Type: "subtitle", Codec: "mov_text", Language: "chi"}}
	if err := db.SaveMedia(context.Background(), item, 1); err != nil {
		t.Fatal(err)
	}
	path := "/api/playback/1/subtitles/-2"
	status(t, request(t, s, "GET", path, nil, nil, nil), 401)
	response := request(t, s, "GET", path, nil, &admin, nil)
	if response.Code != 200 || !strings.HasPrefix(response.Header().Get("Content-Type"), "text/vtt") || !strings.Contains(response.Body.String(), "字幕验证") || response.Header().Get("Cache-Control") != "no-store" {
		t.Fatalf("invalid subtitle response: %d %s", response.Code, response.Body.String())
	}
}
