package server

import (
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"

	"dinosaur/internal/playmetrics"
)

func TestPlaybackMetricSourceAdminOnly(t *testing.T) {
	s, _, admin, root := playbackFixture(t)
	status(t, request(t, s, "POST", "/api/admin/users", credentials{"viewer", testPassword}, &admin, nil), 201)
	viewer := loginAs(t, s, "viewer", testPassword)
	body := metricBody()
	body["media_id"], body["audio_index"], body["compatibility_requested"] = 1, 1, true
	w := request(t, s, "POST", playbackMetricsPost, body, &viewer, nil)
	if w.Code != 204 || w.Body.Len() != 0 {
		t.Fatalf("POST must return no source data: %d %s", w.Code, w.Body)
	}
	w = request(t, s, "GET", playbackMetricsGet, nil, &viewer, nil)
	status(t, w, 403)
	if strings.Contains(w.Body.String(), root) {
		t.Fatal("source path leaked to viewer")
	}
	w = request(t, s, "GET", playbackMetricsGet, nil, &admin, nil)
	status(t, w, 200)
	var result struct {
		Items []playmetrics.Item `json:"items"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil || len(result.Items) != 1 {
		t.Fatalf("invalid response: %s, %v", w.Body, err)
	}
	item := result.Items[0]
	if item.Source == nil || item.Source.Path != filepath.Join(root, "Movie.mp4") || item.Source.AudioCodec != "aac" || item.Source.VideoCodec != "h264" || item.Source.Container != "mp4" || !strings.Contains(item.Reason, "指定 720p") || !strings.Contains(item.Reason, "请求兼容模式") {
		t.Fatalf("incorrect playback source or reason: %s", w.Body)
	}
	body["path"] = "/client/controlled"
	status(t, request(t, s, "POST", playbackMetricsPost, body, &viewer, nil), 400)
}
