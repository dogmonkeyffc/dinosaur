package server

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"
	"time"

	"dinosaur/internal/auth"
	"dinosaur/internal/store"
)

const playbackMetricsPost = "/api/playback/metrics"
const playbackMetricsGet = "/api/admin/playback/metrics"
const playbackMetricID = "abcdefab-0123-4abc-babc-0123456789ab"
const playbackMetricStream = "01234567-89ab-4cde-8f00-0123456789ab"

func metricBody() map[string]any {
	return map[string]any{
		"id": playbackMetricID, "mode": "transcode", "quality": "720", "stream_id": playbackMetricStream,
		"width": 1280, "height": 720, "stalls": 0, "stall_ms": 0, "played_ms": 1000, "buffer_ms": 2000,
	}
}

func TestPlaybackMetricsAPI(t *testing.T) {
	s, db := fixture(t)
	setupAdmin(t, s)
	admin := loginAs(t, s, "admin", testPassword)
	status(t, request(t, s, "POST", "/api/admin/users", credentials{"ordinary", testPassword}, &admin, nil), http.StatusCreated)
	user := loginAs(t, s, "ordinary", testPassword)
	status(t, request(t, s, "POST", playbackMetricsPost, metricBody(), nil, nil), http.StatusUnauthorized)
	status(t, request(t, s, "GET", playbackMetricsGet, nil, nil, nil), http.StatusUnauthorized)
	status(t, request(t, s, "GET", playbackMetricsGet, nil, &user, nil), http.StatusForbidden)
	status(t, request(t, s, "POST", playbackMetricsPost, metricBody(), &user, map[string]string{"X-CSRF-Token": ""}), http.StatusForbidden)
	status(t, request(t, s, "POST", playbackMetricsPost, metricBody(), &user, map[string]string{"Origin": "https://foreign.example"}), http.StatusForbidden)

	empty := request(t, s, "GET", playbackMetricsGet, nil, &admin, nil)
	status(t, empty, http.StatusOK)
	if strings.TrimSpace(empty.Body.String()) != `{"items":[]}` {
		t.Fatalf("expected empty array, got %s", empty.Body)
	}
	ctx := context.Background()
	adminUser, err := db.SessionUser(ctx, auth.TokenHash(admin.cookie.Value))
	if err != nil {
		t.Fatal(err)
	}
	assertDeviceCannotPostMetrics(t, s, db, adminUser)
	if err := db.RecordTranscodeSpeed(ctx, adminUser.ID, playbackMetricStream, 1.25); err != nil {
		t.Fatal(err)
	}

	first := metricBody()
	first["id"] = strings.ToUpper(playbackMetricID)
	first["startup_ms"], first["burn_subtitle"] = 500, true
	status(t, request(t, s, "POST", playbackMetricsPost, first, &admin, nil), http.StatusNoContent)
	update := metricBody()
	update["played_ms"], update["dropped_frames"], update["total_frames"] = 3000, 2, 150
	delete(update, "stream_id")
	status(t, request(t, s, "POST", playbackMetricsPost, update, &admin, nil), http.StatusNoContent)
	status(t, request(t, s, "POST", playbackMetricsPost, metricBody(), &user, nil), http.StatusNoContent)

	w := request(t, s, "GET", playbackMetricsGet, nil, &admin, nil)
	status(t, w, http.StatusOK)
	var result struct {
		Items []map[string]any `json:"items"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil || len(result.Items) != 2 {
		t.Fatalf("expected exactly two items: %s, %v", w.Body, err)
	}
	userItem, adminItem := result.Items[0], result.Items[1]
	for _, item := range result.Items {
		if item["id"] != playbackMetricID || item["updated_at"].(float64) < float64(time.Now().Add(-time.Minute).Unix()) {
			t.Fatalf("unexpected item: %v", item)
		}
		for _, key := range []string{"stream_id", "user_id", "username", "media_id", "title", "path"} {
			if _, ok := item[key]; ok {
				t.Fatalf("unexpected field %s in %v", key, item)
			}
		}
	}
	if userItem["played_ms"] != float64(1000) || userItem["transcode_speed"] != nil || userItem["burn_subtitle"] != nil {
		t.Fatalf("metrics crossed user boundaries: %v", userItem)
	}
	for key, want := range map[string]any{"played_ms": float64(3000), "startup_ms": float64(500), "burn_subtitle": true, "dropped_frames": float64(2), "total_frames": float64(150), "transcode_speed": 1.25} {
		if adminItem[key] != want {
			t.Errorf("%s=%v want %v", key, adminItem[key], want)
		}
	}
}

func assertDeviceCannotPostMetrics(t *testing.T, s *Server, db *store.Store, admin store.User) {
	t.Helper()
	token, err := auth.NewToken()
	if err != nil {
		t.Fatal(err)
	}
	if err := db.CreateSession(context.Background(), auth.TokenHash(token), admin.ID, admin.PasswordHash, time.Now().Add(time.Hour)); err != nil {
		t.Fatal(err)
	}
	if _, err := db.DB.Exec(`UPDATE sessions SET kind='device' WHERE token_hash=?`, auth.TokenHash(token)); err != nil {
		t.Fatal(err)
	}
	status(t, request(t, s, "POST", playbackMetricsPost, metricBody(), nil, map[string]string{"Authorization": "Bearer " + token}), http.StatusForbidden)
}

func TestPlaybackMetricsInvalidBody(t *testing.T) {
	s, _ := fixture(t)
	setupAdmin(t, s)
	admin := loginAs(t, s, "admin", testPassword)
	for key, value := range map[string]any{"played_ms": nil, "height": -1, "width": 16385, "stalls": 0.5, "stream_id": "../title", "id": "00000000-0000-3000-8000-000000000000", "mode": "copy", "quality": "4k", "burn_subtitle": "true", "title": "private"} {
		t.Run(key, func(t *testing.T) {
			body := metricBody()
			body[key] = value
			status(t, request(t, s, "POST", playbackMetricsPost, body, &admin, nil), http.StatusBadRequest)
		})
	}
	missing := metricBody()
	delete(missing, "buffer_ms")
	status(t, request(t, s, "POST", playbackMetricsPost, missing, &admin, nil), http.StatusBadRequest)
}
