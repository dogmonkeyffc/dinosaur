package server

import (
	"encoding/json"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestPlaybackTranscodeSpeedCorrelation(t *testing.T) {
	ffmpeg, err := exec.LookPath("ffmpeg")
	if err != nil {
		t.Skip("FFmpeg is not installed")
	}
	s, _, admin, root := playbackFixture(t)
	path := filepath.Join(root, "Movie.mp4")
	command := exec.Command(ffmpeg, "-y", "-v", "error", "-f", "lavfi", "-i", "color=s=64x48:d=2", "-f", "lavfi", "-i", "anullsrc", "-shortest", "-c:v", "libx264", "-threads", "1", "-c:a", "aac", path)
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("fixture: %s %v", output, err)
	}
	stream := request(t, s, "GET", "/api/playback/1/stream?quality=720&stream_id="+playbackMetricStream, nil, &admin, nil)
	if stream.Code != 200 || stream.Header().Get("Content-Type") != "video/mp4" || stream.Body.Len() == 0 {
		t.Fatalf("expected nonempty MP4 stream: status %d, headers %v", stream.Code, stream.Header())
	}
	status(t, request(t, s, "POST", playbackMetricsPost, metricBody(), &admin, nil), 204)
	w := request(t, s, "GET", playbackMetricsGet, nil, &admin, nil)
	status(t, w, 200)
	var result struct {
		Items []struct {
			Speed float64 `json:"transcode_speed"`
		} `json:"items"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil || len(result.Items) != 1 || result.Items[0].Speed <= 0 {
		t.Fatalf("missing linked FFmpeg speed: %s, %v", w.Body, err)
	}
}
