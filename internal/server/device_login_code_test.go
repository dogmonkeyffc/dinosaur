package server

import (
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"dinosaur/internal/deviceauth"
)

const loginCodePath = "/api/devices/login-code"

func loginCode(t *testing.T, s *Server, who session) string {
	t.Helper()
	w := request(t, s, "POST", loginCodePath, map[string]any{}, &who, nil)
	status(t, w, 201)
	var result struct {
		Code    string `json:"code"`
		Expires int64  `json:"expires_at"`
		Server  string `json:"server_url"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	if !deviceauth.ValidLoginCode(result.Code) || result.Expires < time.Now().Unix()+119 || result.Expires > time.Now().Unix()+120 || result.Server != "http://example.com" {
		t.Fatalf("invalid code response: %s", w.Body.String())
	}
	return result.Code
}

func codeExchange(t *testing.T, s *Server, code string) *httptest.ResponseRecorder {
	t.Helper()
	return request(t, s, "POST", loginCodePath+"/exchange", map[string]string{"code": code, "name": "手机"}, nil, nil)
}

func TestLoginCodeAuthenticationAndOwner(t *testing.T) {
	s, db := fixture(t)
	setupAdmin(t, s)
	admin := loginAs(t, s, "admin", testPassword)
	status(t, request(t, s, "POST", "/api/admin/users", credentials{"viewer", testPassword}, &admin, nil), 201)
	viewer := loginAs(t, s, "viewer", testPassword)
	status(t, request(t, s, "POST", loginCodePath, map[string]any{}, nil, nil), 401)
	noCSRF := session{cookie: viewer.cookie}
	status(t, request(t, s, "POST", loginCodePath, map[string]any{}, &noCSRF, nil), 403)
	code := loginCode(t, s, viewer)
	// A different user's refresh must not replace this owner's credential.
	loginCode(t, s, admin)
	var hash string
	if err := db.DB.QueryRow(`SELECT secret_hash FROM device_grants WHERE owner_hash=(SELECT token_hash FROM sessions WHERE user_id=(SELECT id FROM users WHERE username='viewer'))`).Scan(&hash); err != nil {
		t.Fatal(err)
	}
	if len(hash) != 64 || strings.Contains(hash, code) {
		t.Fatal("code not stored as hash")
	}
	w := codeExchange(t, s, code)
	status(t, w, 200)
	var native deviceauth.Session
	if err := json.Unmarshal(w.Body.Bytes(), &native); err != nil {
		t.Fatal(err)
	}
	if native.User.Username != "viewer" || len(native.Token) != 43 || len(w.Result().Cookies()) != 0 {
		t.Fatal("wrong device identity")
	}
	status(t, request(t, s, "GET", "/api/auth/me", nil, nil, bearer(native.Token)), 200)
	status(t, request(t, s, "POST", loginCodePath, map[string]any{}, nil, bearer(native.Token)), 403)
	status(t, request(t, s, "POST", "/api/devices/qr", map[string]any{}, nil, bearer(native.Token)), 403)
	status(t, codeExchange(t, s, code), 410)
}

func TestLoginCodeReplacementExpiryAndLogout(t *testing.T) {
	s, db := fixture(t)
	setupAdmin(t, s)
	who := loginAs(t, s, "admin", testPassword)
	old := loginCode(t, s, who)
	fresh := loginCode(t, s, who)
	status(t, codeExchange(t, s, old), 410)
	ticket := mobileTicket(t, s, who)
	status(t, codeExchange(t, s, fresh), 410)
	fresh = loginCode(t, s, who)
	status(t, request(t, s, "POST", "/api/devices/qr/exchange", map[string]string{"ticket": ticket, "name": "手机"}, nil, nil), 410)
	if _, err := db.DB.Exec(`UPDATE device_grants SET expires_at=?`, time.Now().Unix()); err != nil {
		t.Fatal(err)
	}
	status(t, codeExchange(t, s, fresh), 410)
	fresh = loginCode(t, s, who)
	status(t, request(t, s, "POST", "/api/auth/logout", nil, &who, nil), 200)
	status(t, codeExchange(t, s, fresh), 410)
}

func TestLoginCodeBadInputAndRate(t *testing.T) {
	s, _ := fixture(t)
	setupAdmin(t, s)
	who := loginAs(t, s, "admin", testPassword)
	status(t, request(t, s, "POST", loginCodePath, map[string]string{"unexpected": "x"}, &who, nil), 400)
	for _, code := range []string{"", "123456", "１２３４５６７８", "1234567x", strings.Repeat("a", 43)} {
		status(t, codeExchange(t, s, code), 400)
	}
	for _, body := range []any{map[string]any{"code": 12345678, "name": "phone"}, map[string]string{"code": "12345678", "name": "\n"}, map[string]string{"code": "12345678", "name": "phone", "ticket": "x"}} {
		status(t, request(t, s, "POST", loginCodePath+"/exchange", body, nil, nil), 400)
	}
	for i := 0; i < 2; i++ {
		status(t, codeExchange(t, s, "12345678"), 410)
	}
	w := codeExchange(t, s, "12345678")
	status(t, w, 429)
	if w.Header().Get("Retry-After") != "900" {
		t.Fatal("missing retry duration")
	}
}

func TestLoginCodeGlobalGuessBudget(t *testing.T) {
	s, _ := fixture(t)
	for i := 0; i < 101; i++ {
		r := httptest.NewRequest("POST", "http://example.com"+loginCodePath+"/exchange", strings.NewReader(`{"code":"12345678","name":"phone"}`))
		r.Header.Set("Content-Type", "application/json")
		r.RemoteAddr = "peer-" + strings.Repeat("x", i) + ":1000"
		w := httptest.NewRecorder()
		s.Handler().ServeHTTP(w, r)
		want := 410
		if i == 100 {
			want = 429
		}
		status(t, w, want)
	}
}
