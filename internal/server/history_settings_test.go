package server

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"dinosaur/internal/history"
	"dinosaur/internal/store"
)

func TestHistorySettingAuthenticationAndAccountIsolation(t *testing.T) {
	s, db := fixture(t)
	setupAdmin(t, s)
	admin := loginAs(t, s, "admin", testPassword)
	u, _ := db.UserByName(context.Background(), "admin")
	if _, err := db.CreateUser(context.Background(), "viewer", u.PasswordHash, "user", false); err != nil {
		t.Fatal(err)
	}
	viewer := loginAs(t, s, "viewer", testPassword)
	lib, err := db.SaveLibrary(context.Background(), store.Library{Name: "Movies", Type: "movie", RootPath: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	path := fmt.Sprintf("/api/history/settings/%d", lib.ID)
	body := map[string]bool{"record_history": false}
	status(t, request(t, s, "GET", "/api/history/settings", nil, nil, nil), 401)
	status(t, request(t, s, "PUT", path, body, &admin, map[string]string{"X-CSRF-Token": ""}), 403)
	status(t, request(t, s, "PUT", path, map[string]any{}, &admin, nil), 400)
	status(t, request(t, s, "PUT", path, map[string]any{"record_history": false, "user_id": 2}, &admin, nil), 400)
	status(t, request(t, s, "PUT", path, body, &admin, nil), 200)
	for _, test := range []struct {
		who     session
		enabled bool
	}{{admin, false}, {viewer, true}} {
		w := request(t, s, "GET", "/api/history/settings", nil, &test.who, nil)
		status(t, w, 200)
		var result struct {
			Libraries []history.LibrarySetting `json:"libraries"`
		}
		if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil || len(result.Libraries) != 1 || result.Libraries[0].RecordHistory != test.enabled {
			t.Fatalf("bad setting: %s %v", w.Body, err)
		}
	}
	native := exchangeMobile(t, s, mobileTicket(t, s, viewer))
	status(t, request(t, s, "PUT", path, body, nil, bearer(native.Token)), 200)
	if _, err := db.DB.Exec(`UPDATE libraries SET restricted=1 WHERE id=?`, lib.ID); err != nil {
		t.Fatal(err)
	}
	status(t, request(t, s, "PUT", path, body, &viewer, nil), 404)
	status(t, request(t, s, "PUT", "/api/history/settings/9999", body, &admin, nil), 404)
}
