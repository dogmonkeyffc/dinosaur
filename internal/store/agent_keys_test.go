package store

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"dinosaur/internal/agent"
	"dinosaur/internal/auth"
	"dinosaur/internal/metadata"
)

func TestAgentKeyRevocationExpiryOwnerAndLibraryScope(t *testing.T) {
	db, key, tvLibrary := agentStoreFixture(t)
	ctx := context.Background()
	got, err := db.AuthenticateAgentKey(ctx, auth.TokenHash("test-agent-secret"))
	if err != nil || got.ID != key.ID {
		t.Fatal("key authentication failed")
	}
	service := agent.Service{Repository: db}
	scoped := key
	scoped.LibraryID = tvLibrary
	page, err := service.Tasks(ctx, scoped, metadata.ReportFilter{Page: 1, PageSize: 20, Status: "pending"})
	if err != nil || page.Total != 1 || page.Items[0].SeriesID != 1 {
		t.Fatalf("scope leaked: %+v %v", page, err)
	}
	if _, err := service.Tasks(ctx, scoped, metadata.ReportFilter{LibraryID: 1, Page: 1, PageSize: 20}); !errors.Is(err, agent.ErrForbidden) {
		t.Fatal("explicit scope bypass")
	}
	readonly := key
	readonly.Access = "read"
	if _, err := service.Update(ctx, readonly, "movie", 1, 0, agent.Input{Title: "bad"}, "agent"); !errors.Is(err, agent.ErrForbidden) {
		t.Fatal("read-only key wrote metadata")
	}
	for _, query := range []string{`UPDATE users SET disabled=1 WHERE id=?`, `UPDATE users SET disabled=0,role='user' WHERE id=?`} {
		if _, err := db.DB.Exec(query, key.UserID); err != nil {
			t.Fatal(err)
		}
		if _, err := db.AuthenticateAgentKey(ctx, auth.TokenHash("test-agent-secret")); !errors.Is(err, sql.ErrNoRows) {
			t.Fatal("disabled/demoted owner still authenticates")
		}
	}
	if _, err := db.DB.Exec(`UPDATE users SET role='admin' WHERE id=?`, key.UserID); err != nil {
		t.Fatal(err)
	}
	if _, err := db.DB.Exec(`UPDATE agent_keys SET expires_at=? WHERE id=?`, time.Now().Add(-time.Hour).Unix(), key.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := db.AuthenticateAgentKey(ctx, auth.TokenHash("test-agent-secret")); !errors.Is(err, sql.ErrNoRows) {
		t.Fatal("expired key accepted")
	}
	if err := db.RevokeAgentKey(ctx, key.UserID, key.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := db.AuthenticateAgentKey(ctx, auth.TokenHash("test-agent-secret")); !errors.Is(err, sql.ErrNoRows) {
		t.Fatal("revoked key accepted")
	}
	key.LibraryID = tvLibrary
	key.ExpiresAt = time.Now().Add(time.Hour).Unix()
	scopedKey, err := db.CreateAgentKey(ctx, key, auth.TokenHash("scoped-secret"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.DB.Exec(`DELETE FROM libraries WHERE id=?`, tvLibrary); err != nil {
		t.Fatal(err)
	}
	if _, err := db.AuthenticateAgentKey(ctx, auth.TokenHash("scoped-secret")); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("deleted scope became an all-library key: %d", scopedKey.ID)
	}
}
