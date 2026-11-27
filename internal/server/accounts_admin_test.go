package server

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"dinosaur/internal/store"
)

func TestAdministratorResetsPasswordAndRevokesSessions(t *testing.T) {
	s, db := fixture(t)
	setupAdmin(t, s)
	admin := loginAs(t, s, "admin", testPassword)
	user := createManagedAccount(t, s, db, "reset-viewer", admin)
	viewer := loginAs(t, s, user.Username, testPassword)
	phone := exchangeMobile(t, s, mobileTicket(t, s, viewer))
	path := fmt.Sprintf("/api/admin/users/%d/password", user.ID)
	newPassword := "reset password for test"
	body := map[string]string{"new_password": newPassword}
	status(t, request(t, s, "PUT", path, body, &viewer, nil), 403)
	status(t, request(t, s, "PUT", path, body, &admin, map[string]string{"X-CSRF-Token": ""}), 403)
	status(t, request(t, s, "PUT", path, map[string]string{"new_password": "short"}, &admin, nil), 400)
	status(t, request(t, s, "PUT", "/api/admin/users/1/password", body, &admin, nil), 409)
	status(t, request(t, s, "PUT", "/api/admin/users/99999/password", body, &admin, nil), 404)
	status(t, request(t, s, "PUT", path, body, &admin, nil), 204)
	status(t, request(t, s, "GET", "/api/auth/me", nil, &viewer, nil), 401)
	status(t, request(t, s, "GET", "/api/auth/me", nil, nil, bearer(phone.Token)), 401)
	status(t, request(t, s, "GET", "/api/auth/me", nil, &admin, nil), 200)
	status(t, request(t, s, "POST", "/api/auth/login", credentials{user.Username, testPassword}, nil, nil), 401)
	viewer = loginAs(t, s, user.Username, newPassword)
	status(t, request(t, s, "GET", "/api/auth/me", nil, &viewer, nil), 200)
}

func TestAccountLibraryChangesPreserveOtherGrants(t *testing.T) {
	s, db := fixture(t)
	setupAdmin(t, s)
	admin := loginAs(t, s, "admin", testPassword)
	user := createManagedAccount(t, s, db, "library-viewer", admin)
	other := createManagedAccount(t, s, db, "other-viewer", admin)
	viewer := loginAs(t, s, user.Username, testPassword)
	secondary := loginAs(t, s, other.Username, testPassword)
	ctx := context.Background()
	makeLibrary := func(name string) int64 {
		t.Helper()
		lib, err := db.SaveLibrary(ctx, store.Library{Name: name, Type: "video", RootPath: t.TempDir()})
		if err != nil {
			t.Fatal(err)
		}
		return lib.ID
	}
	private, public := makeLibrary("Private"), makeLibrary("Public")
	if err := db.SaveLibraryAccess(ctx, private, store.LibraryAccess{Restricted: true, UserIDs: []int64{other.ID}}); err != nil {
		t.Fatal(err)
	}
	path := fmt.Sprintf("/api/admin/users/%d/libraries", user.ID)
	rule := func(id int64, grant bool) map[string]any {
		return map[string]any{"library_id": id, "granted": grant}
	}
	change := func(code int, rules ...map[string]any) {
		t.Helper()
		status(t, request(t, s, "PATCH", path, map[string]any{"changes": rules}, &admin, nil), code)
	}
	publicPath, privatePath := fmt.Sprintf("/api/libraries/%d/videos", public), fmt.Sprintf("/api/libraries/%d/videos", private)
	status(t, request(t, s, "GET", path, nil, &viewer, nil), 403)
	status(t, request(t, s, "PATCH", path, map[string]any{"changes": []any{rule(private, true)}}, &viewer, nil), 403)
	status(t, request(t, s, "GET", "/api/admin/users/1/libraries", nil, &admin, nil), 409)
	status(t, request(t, s, "GET", publicPath, nil, &viewer, nil), 200)

	checkGrant := func(granted bool) {
		t.Helper()
		w := request(t, s, "GET", path, nil, &admin, nil)
		status(t, w, 200)
		var response struct{ Libraries []store.UserLibrary }
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil || len(response.Libraries) != 2 {
			t.Fatalf("invalid access state: %s", w.Body)
		}
		first, second := response.Libraries[0], response.Libraries[1]
		if first.ID != private || !first.Restricted || first.Granted != granted || second.ID != public || second.Restricted {
			t.Fatalf("unexpected access state: %s", w.Body)
		}
	}
	change(400, rule(private, true), rule(private, false))
	change(409, rule(private, true), rule(public, false))
	checkGrant(false)
	status(t, request(t, s, "GET", privatePath, nil, &viewer, nil), 404)
	change(204, rule(private, true))
	change(204, rule(private, true))
	checkGrant(true)
	status(t, request(t, s, "GET", privatePath, nil, &viewer, nil), 200)
	change(204, rule(private, false))
	checkGrant(false)
	status(t, request(t, s, "GET", privatePath, nil, &viewer, nil), 404)
	status(t, request(t, s, "GET", privatePath, nil, &secondary, nil), 200)
}

func createManagedAccount(t *testing.T, s *Server, db *store.Store, name string, admin session) store.User {
	t.Helper()
	status(t, request(t, s, "POST", "/api/admin/users", credentials{name, testPassword}, &admin, nil), 201)
	user, err := db.UserByName(context.Background(), name)
	if err != nil {
		t.Fatal(err)
	}
	return user
}
