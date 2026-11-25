package server

import (
	"context"
	"encoding/json"
	"testing"

	"dinosaur/internal/playback"
)

func TestDisabledRecordingReturnsNoProgress(t *testing.T) {
	s, db, admin, _ := playbackFixture(t)
	user, err := db.UserByName(context.Background(), "admin")
	if err != nil {
		t.Fatal(err)
	}
	status(t, request(t, s, "PUT", "/api/playback/1/progress", map[string]any{"position": 120}, &admin, nil), 200)
	status(t, request(t, s, "PUT", "/api/history/settings/1", map[string]bool{"record_history": false}, &admin, nil), 200)
	for _, method := range []string{"GET", "PUT"} {
		path := "/api/playback/1"
		var body any
		if method == "PUT" {
			path += "/progress"
			body = map[string]any{"position": 900, "watched": true}
		}
		w := request(t, s, method, path, body, &admin, nil)
		status(t, w, 200)
		var result struct {
			Progress playback.Progress `json:"progress"`
			Resume   float64           `json:"resume_position"`
		}
		if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
			t.Fatal(err)
		}
		if result.Progress.Position != 0 || result.Progress.Watched || result.Resume != 0 {
			t.Fatalf("progress leaked in %s: %s", method, w.Body)
		}
	}
	var stored int
	if err := db.DB.QueryRow(`SELECT count(*) FROM playback_progress WHERE user_id=? AND media_id=1`, user.ID).Scan(&stored); err != nil || stored != 0 {
		t.Fatal("disabled progress retained", stored, err)
	}
}
