package server

import (
	"context"
	"strings"
	"testing"

	"dinosaur/internal/agent"
	"dinosaur/internal/auth"
)

func TestAgentKeysAreDedicatedScopedRevocableCredentials(t *testing.T) {
	s, db, admin, movieLibrary, tvLibrary := agentFixture(t)
	key, token := createTestAgentKey(t, s, &admin, movieLibrary, "read_write")
	listed := request(t, s, "GET", "/api/admin/agent-keys", nil, &admin, nil)
	status(t, listed, 200)
	if strings.Contains(listed.Body.String(), token) || strings.Contains(listed.Body.String(), auth.TokenHash(token)) {
		t.Fatal("key secret leaked in listing")
	}
	var stored string
	if err := db.DB.QueryRow(`SELECT token_hash FROM agent_keys WHERE id=?`, key.ID).Scan(&stored); err != nil || stored != auth.TokenHash(token) {
		t.Fatal("raw key stored instead of hash")
	}
	status(t, agentRequest(t, s, "GET", "/api/agent/tasks", nil, "", ""), 401)
	status(t, request(t, s, "GET", "/api/agent/tasks", nil, &admin, nil), 401)
	status(t, request(t, s, "GET", "/api/agent/tasks", nil, &admin, map[string]string{"Authorization": "Bearer " + token}), 401)
	status(t, agentRequest(t, s, "GET", "/api/admin/users", nil, token, ""), 401)
	status(t, agentRequest(t, s, "GET", "/api/media/1", nil, token, ""), 401)
	status(t, agentRequest(t, s, "GET", "/api/agent/movies/1", nil, token, ""), 200)
	status(t, agentRequest(t, s, "GET", "/api/agent/series/1", nil, token, ""), 403)
	status(t, agentRequest(t, s, "GET", "/api/agent/tasks?library_id="+stringID(tvLibrary), nil, token, ""), 403)
	result := agentRequest(t, s, "GET", "/api/agent/tasks", nil, token, "")
	status(t, result, 200)
	if strings.Contains(result.Body.String(), `"library_name":"TV"`) {
		t.Fatal("scoped tasks exposed another library")
	}
	_, readonly := createTestAgentKey(t, s, &admin, 0, "read")
	status(t, agentRequest(t, s, "PUT", "/api/agent/movies/1/metadata", agent.Input{Title: "Not allowed"}, readonly, `"0"`), 403)
	status(t, request(t, s, "DELETE", "/api/admin/agent-keys/"+stringID(key.ID), nil, &admin, map[string]string{"X-CSRF-Token": ""}), 403)
	status(t, request(t, s, "DELETE", "/api/admin/agent-keys/"+stringID(key.ID), nil, &admin, nil), 200)
	status(t, agentRequest(t, s, "GET", "/api/agent/tasks", nil, token, ""), 401)
	hash, err := s.hasher.Hash(context.Background(), testPassword)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.CreateUser(context.Background(), "viewer", hash, "user", false); err != nil {
		t.Fatal(err)
	}
	viewer := loginAs(t, s, "viewer", testPassword)
	status(t, request(t, s, "POST", "/api/admin/agent-keys", agent.CreateKey{Name: "bad"}, &viewer, nil), 403)
	status(t, request(t, s, "GET", "/api/admin/agent-keys", nil, &viewer, nil), 403)
	for _, bad := range []agent.CreateKey{{Name: ""}, {Name: "Name", LibraryID: -1}, {Name: "Name", Access: "admin"}, {Name: "Name", ExpiresInDays: 366}} {
		status(t, request(t, s, "POST", "/api/admin/agent-keys", bad, &admin, nil), 400)
	}
}
