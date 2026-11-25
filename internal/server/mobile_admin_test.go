package server

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"dinosaur/internal/auth"
)

func adminCode(t *testing.T, s *Server, token string) string {
	t.Helper()
	w := request(t, s, "POST", "/api/devices/admin-web", map[string]any{}, nil, bearer(token))
	status(t, w, 200)
	var result struct {
		Code string `json:"code"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil || len(result.Code) != 43 {
		t.Fatalf("invalid grant %v", err)
	}
	return result.Code
}

func exchangeAdmin(t *testing.T, s *Server, code string) *httptest.ResponseRecorder {
	t.Helper()
	r := httptest.NewRequest("POST", "http://example.com/api/devices/admin-web/exchange", strings.NewReader(url.Values{"code": {code}}.Encode()))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	s.Handler().ServeHTTP(w, r)
	return w
}

func TestMobileAdminBrowserSessionIsScopedAndRevocable(t *testing.T) {
	s, db := fixture(t)
	setupAdmin(t, s)
	admin := loginAs(t, s, "admin", testPassword)
	native := exchangeMobile(t, s, mobileTicket(t, s, admin))
	status(t, request(t, s, "POST", "/api/devices/admin-web", map[string]any{}, nil, nil), 401)
	status(t, request(t, s, "POST", "/api/devices/admin-web", map[string]any{}, &admin, nil), 403)
	user, err := db.UserByName(context.Background(), "admin")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.CreateUser(context.Background(), "viewer", user.PasswordHash, "user", false); err != nil {
		t.Fatal(err)
	}
	viewer := loginAs(t, s, "viewer", testPassword)
	ordinary := exchangeMobile(t, s, mobileTicket(t, s, viewer))
	status(t, request(t, s, "POST", "/api/devices/admin-web", map[string]any{}, nil, bearer(ordinary.Token)), 403)
	code := adminCode(t, s, native.Token)
	status(t, request(t, s, "POST", "/api/devices/admin-web/exchange", map[string]string{"code": code}, nil, nil), 415)
	w := exchangeAdmin(t, s, code)
	status(t, w, 303)
	if w.Header().Get("Location") != "/settings" || strings.Contains(w.Header().Get("Location"), code) {
		t.Fatal("unsafe redirect")
	}
	cookies := w.Result().Cookies()
	if len(cookies) != 1 || !cookies[0].HttpOnly {
		t.Fatal("missing HttpOnly session")
	}
	browser := session{cookie: cookies[0], csrf: auth.CSRFToken(cookies[0].Value)}
	status(t, request(t, s, "GET", "/api/admin/users", nil, &browser, nil), 200)
	status(t, request(t, s, "POST", "/api/admin/users", map[string]string{"username": "created", "password": testPassword, "role": "user"}, &browser, nil), 201)
	status(t, request(t, s, "POST", "/api/admin/users", map[string]string{}, &browser, map[string]string{"X-CSRF-Token": ""}), 403)
	status(t, exchangeAdmin(t, s, code), 401)
	// Reopening revokes the previous embedded browser without logging out the phone.
	nextCode := adminCode(t, s, native.Token)
	status(t, request(t, s, "GET", "/api/admin/users", nil, &browser, nil), 401)
	w = exchangeAdmin(t, s, nextCode)
	status(t, w, 303)
	browser.cookie = w.Result().Cookies()[0]
	browser.csrf = auth.CSRFToken(browser.cookie.Value)
	status(t, request(t, s, "POST", "/api/devices/admin-web/close", map[string]any{}, nil, bearer(native.Token)), 200)
	status(t, request(t, s, "GET", "/api/admin/users", nil, &browser, nil), 401)
	w = exchangeAdmin(t, s, adminCode(t, s, native.Token))
	status(t, w, 303)
	browser.cookie = w.Result().Cookies()[0]
	if err := db.DeleteSession(context.Background(), auth.TokenHash(native.Token)); err != nil {
		t.Fatal(err)
	}
	status(t, request(t, s, "GET", "/api/admin/users", nil, &browser, nil), 401)
	status(t, request(t, s, "GET", "/api/admin/users", nil, &admin, nil), 200)
}

func TestMobileAdminGrantExpiresAndCannotOutliveDevice(t *testing.T) {
	s, db := fixture(t)
	setupAdmin(t, s)
	admin := loginAs(t, s, "admin", testPassword)
	native := exchangeMobile(t, s, mobileTicket(t, s, admin))
	code := adminCode(t, s, native.Token)
	if _, err := db.DB.Exec(`UPDATE mobile_admin_grants SET expires_at=0`); err != nil {
		t.Fatal(err)
	}
	status(t, exchangeAdmin(t, s, code), 401)
	deadline := time.Now().Add(30 * time.Second).Unix()
	if _, err := db.DB.Exec(`UPDATE sessions SET expires_at=? WHERE token_hash=?`, deadline, auth.TokenHash(native.Token)); err != nil {
		t.Fatal(err)
	}
	w := exchangeAdmin(t, s, adminCode(t, s, native.Token))
	status(t, w, 303)
	if w.Result().Cookies()[0].Expires.Unix() > deadline {
		t.Fatal("embedded browser outlives its device")
	}
	code = adminCode(t, s, native.Token)
	if err := db.DeleteSession(context.Background(), auth.TokenHash(native.Token)); err != nil {
		t.Fatal(err)
	}
	status(t, exchangeAdmin(t, s, code), 401)
}
