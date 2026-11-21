package server

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"dinosaur/internal/television"
)

func TestAgentImportsUnrecognizedEpisodeAndKeepsAssignmentOnRescan(t *testing.T) {
	s, db, admin, movie, tv := agentFixture(t)
	ctx := context.Background()
	library, err := db.LibraryByID(ctx, tv)
	if err != nil {
		t.Fatal(err)
	}
	filename := "New Show/Season 02/[01].mkv"
	target := filepath.Join(library.RootPath, filename)
	if err := os.MkdirAll(filepath.Dir(target), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(target, []byte("test fixture"), 0600); err != nil {
		t.Fatal(err)
	}
	_, key := createTestAgentKey(t, s, &admin, tv, "read_write")
	_, readKey := createTestAgentKey(t, s, &admin, tv, "read")
	_, otherKey := createTestAgentKey(t, s, &admin, movie, "read_write")
	path := "/api/agent/libraries/" + stringID(tv)
	body := map[string]any{"episodes": []television.EpisodeAssignment{{Path: filename, SeriesFolder: "New Show", Season: 2, Episode: 1}}}
	for _, unauthorized := range []string{readKey, otherKey} {
		status(t, agentRequest(t, s, "POST", path+"/episodes", body, unauthorized, ""), 403)
	}
	invalid := map[string]any{"episodes": []television.EpisodeAssignment{{Path: "../[01].mkv", SeriesFolder: "..", Season: 2, Episode: 1}}}
	status(t, agentRequest(t, s, "POST", path+"/episodes", invalid, key, ""), 400)
	for range 2 {
		status(t, agentRequest(t, s, "POST", path+"/episodes", body, key, ""), 202)
		deadline := time.Now().Add(8 * time.Second)
		finished := false
		for time.Now().Before(deadline) {
			scan, err := db.LatestScan(ctx, tv)
			if err != nil {
				t.Fatal(err)
			}
			if scan != nil && scan.Status == "completed" {
				finished = true
				break
			}
			time.Sleep(20 * time.Millisecond)
		}
		if !finished {
			t.Fatal("scan did not complete")
		}
		status(t, agentRequest(t, s, "GET", path+"/scan", nil, readKey, ""), 200)
		var season, episode, missing int
		err = db.DB.QueryRow(`SELECT e.season_number,e.episode_number,m.missing FROM tv_episodes e JOIN media_items m ON m.id=e.media_id WHERE m.library_id=? AND m.relative_path=?`,
			tv, filename).Scan(&season, &episode, &missing)
		if err != nil || season != 2 || episode != 1 || missing != 0 {
			t.Fatalf("episode status: season=%d episode=%d missing=%d err=%v", season, episode, missing, err)
		}
	}
	body["episodes"] = []television.EpisodeAssignment{{Path: filename, SeriesFolder: "New Show", Season: 2, Episode: 2}}
	status(t, agentRequest(t, s, "POST", path+"/episodes", body, key, ""), 409)
}
