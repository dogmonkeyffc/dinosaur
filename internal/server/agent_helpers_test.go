package server

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http/httptest"
	"strconv"
	"testing"

	"dinosaur/internal/agent"
	"dinosaur/internal/media"
	"dinosaur/internal/store"
	"dinosaur/internal/television"
)

func agentFixture(t *testing.T) (*Server, *store.Store, session, int64, int64) {
	t.Helper()
	s, db := fixture(t)
	setupAdmin(t, s)
	admin := loginAs(t, s, "admin", testPassword)
	ctx := context.Background()
	movie, err := db.SaveLibrary(ctx, store.Library{Name: "Movies", Type: "movie", RootPath: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.SaveMedia(ctx, media.Item{LibraryID: movie.ID, Title: "Movie", Year: 1999, RelativePath: "Movie.1999.mkv"}, 1); err != nil {
		t.Fatal(err)
	}
	tv, err := db.SaveLibrary(ctx, store.Library{Name: "TV", Type: "tv", RootPath: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{"Show (2024)/Show.S01E01.mkv", "Show (2024)/Show.S01E02.mkv"} {
		name, _ := television.ParseEpisode(path)
		if err := db.SaveEpisode(ctx, media.Item{LibraryID: tv.ID, RelativePath: path, Title: name.Title}, name, 1); err != nil {
			t.Fatal(err)
		}
	}
	return s, db, admin, movie.ID, tv.ID
}

func createTestAgentKey(t *testing.T, s *Server, admin *session, library int64, access string) (agent.Key, string) {
	t.Helper()
	result := request(t, s, "POST", "/api/admin/agent-keys", agent.CreateKey{Name: "Test Agent", LibraryID: library, Access: access, ExpiresInDays: 90}, admin, nil)
	status(t, result, 201)
	var response struct {
		Key   agent.Key `json:"key"`
		Token string    `json:"token"`
	}
	if err := json.Unmarshal(result.Body.Bytes(), &response); err != nil {
		t.Fatal(err)
	}
	if len(response.Token) != len(agentTokenPrefix)+43 {
		t.Fatal("invalid key format")
	}
	return response.Key, response.Token
}

func agentRequest(t *testing.T, s *Server, method, path string, body any, token, version string) *httptest.ResponseRecorder {
	t.Helper()
	var data []byte
	if body != nil {
		var err error
		data, err = json.Marshal(body)
		if err != nil {
			t.Fatal(err)
		}
	}
	return agentRaw(t, s, method, path, data, "application/json", token, version)
}

func agentRaw(t *testing.T, s *Server, method, path string, data []byte, content, token, version string) *httptest.ResponseRecorder {
	t.Helper()
	r := httptest.NewRequest(method, "http://example.com"+path, bytes.NewReader(data))
	if content != "" {
		r.Header.Set("Content-Type", content)
	}
	if token != "" {
		r.Header.Set("Authorization", "Bearer "+token)
	}
	if version != "" {
		r.Header.Set("If-Match", version)
	}
	result := httptest.NewRecorder()
	s.Handler().ServeHTTP(result, r)
	return result
}

func stringID(id int64) string { return strconv.FormatInt(id, 10) }
