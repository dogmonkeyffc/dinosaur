package server

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"dinosaur/internal/auth"
)

func TestAuthLifecycle(t *testing.T) {
	s, db := fixture(t)
	status(t, request(t, s, "GET", "/api/setup", nil, nil, nil), 200)
	status(t, request(t, s, "POST", "/api/setup", credentials{"admin", testPassword}, nil, nil), 403)
	status(t, request(t, s, "POST", "/api/setup", credentials{"admin", "short"}, nil, map[string]string{"X-Setup-Token": setupSecret}), 400)
	setupAdmin(t, s)
	status(t, request(t, s, "POST", "/api/setup", credentials{"other", testPassword}, nil, nil), 409)
	unknown := request(t, s, "POST", "/api/auth/login", credentials{"unknown", testPassword}, nil, nil)
	wrong := request(t, s, "POST", "/api/auth/login", credentials{"admin", "wrong"}, nil, nil)
	status(t, unknown, 401)
	status(t, wrong, 401)
	if unknown.Body.String() != wrong.Body.String() {
		t.Fatal("credential errors reveal account existence")
	}
	admin := loginAs(t, s, "ADMIN", testPassword)
	status(t, request(t, s, "GET", "/api/auth/me", nil, &admin, nil), 200)
	for _, path := range []string{"/api/media/1", "/api/playback/1", "/api/media/1/poster", "/api/media/1/subtitles/en"} {
		status(t, request(t, s, "GET", path, nil, nil, nil), 401)
		want := 501
		if path == "/api/media/1" || path == "/api/media/1/poster" || path == "/api/playback/1" {
			want = 404
		}
		status(t, request(t, s, "GET", path, nil, &admin, nil), want)
	}
	userBody := map[string]string{"username": "viewer", "password": testPassword}
	status(t, request(t, s, "POST", "/api/admin/users", userBody, &admin, map[string]string{"X-CSRF-Token": ""}), 403)
	status(t, request(t, s, "POST", "/api/admin/users", userBody, &admin, map[string]string{"Origin": "https://attacker.example"}), 403)
	status(t, request(t, s, "POST", "/api/admin/users", userBody, &admin, map[string]string{"Origin": "null"}), 403)
	status(t, request(t, s, "POST", "/api/admin/users", userBody, &admin, nil), 201)
	status(t, request(t, s, "POST", "/api/admin/users", userBody, &admin, nil), 409)
	viewer := loginAs(t, s, "viewer", testPassword)
	viewer2 := loginAs(t, s, "viewer", testPassword)
	status(t, request(t, s, "POST", "/api/admin/users", userBody, &viewer, nil), 403)
	status(t, request(t, s, "POST", "/api/admin/libraries/scan", map[string]string{}, &viewer, nil), 403)
	status(t, request(t, s, "GET", "/api/media/1", nil, &viewer, nil), 404)
	status(t, request(t, s, "POST", "/api/auth/password", map[string]string{"current_password": "wrong", "new_password": "another valid password"}, &viewer, nil), 403)
	status(t, request(t, s, "POST", "/api/auth/password", map[string]string{"current_password": testPassword, "new_password": "another valid password"}, &viewer, nil), 200)
	status(t, request(t, s, "GET", "/api/auth/me", nil, &viewer, nil), 401)
	status(t, request(t, s, "GET", "/api/auth/me", nil, &viewer2, nil), 401)
	status(t, request(t, s, "POST", "/api/auth/login", credentials{"viewer", testPassword}, nil, nil), 401)
	viewer = loginAs(t, s, "viewer", "another valid password")
	status(t, request(t, s, "POST", "/api/auth/logout", nil, &viewer, nil), 200)
	status(t, request(t, s, "GET", "/api/auth/me", nil, &viewer, nil), 401)
	viewer = loginAs(t, s, "viewer", "another valid password")
	u, err := db.UserByName(context.Background(), "viewer")
	if err != nil {
		t.Fatal(err)
	}
	path := fmt.Sprintf("/api/admin/users/%d", u.ID)
	status(t, request(t, s, "PATCH", path, map[string]bool{"disabled": true}, &admin, nil), 200)
	status(t, request(t, s, "GET", "/api/auth/me", nil, &viewer, nil), 401)
	status(t, request(t, s, "POST", "/api/auth/login", credentials{"viewer", "another valid password"}, nil, nil), 401)
	status(t, request(t, s, "PATCH", path, map[string]bool{"disabled": false}, &admin, nil), 200)
	status(t, request(t, s, "GET", "/api/auth/me", nil, &viewer, nil), 401)
	viewer = loginAs(t, s, "viewer", "another valid password")
	status(t, request(t, s, "PATCH", "/api/admin/users/1", map[string]bool{"disabled": true}, &admin, nil), 409)
	if _, err := db.DB.Exec(`UPDATE sessions SET expires_at=0 WHERE token_hash=?`, auth.TokenHash(viewer.cookie.Value)); err != nil {
		t.Fatal(err)
	}
	status(t, request(t, s, "GET", "/api/auth/me", nil, &viewer, nil), 401)
	var raw string
	if err := db.DB.QueryRow(`SELECT token_hash FROM sessions WHERE user_id=1`).Scan(&raw); err != nil {
		t.Fatal(err)
	}
	if raw == admin.cookie.Value || len(raw) != 64 {
		t.Fatal("raw session persisted")
	}
}
func TestConcurrentSetup(t *testing.T) {
	s, db := fixture(t)
	const n = 4
	results := make(chan int, n)
	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			w := request(t, s, "POST", "/api/setup", credentials{"admin", testPassword}, nil, map[string]string{"X-Setup-Token": setupSecret})
			results <- w.Code
		}()
	}
	wg.Wait()
	close(results)
	created := 0
	for code := range results {
		if code == 201 {
			created++
		} else if code != 409 {
			t.Fatalf("unexpected status %d", code)
		}
	}
	if created != 1 {
		t.Fatalf("created %d administrators", created)
	}
	var count int
	if err := db.DB.QueryRow(`SELECT count(*) FROM users`).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatalf("count=%d", count)
	}
}
func TestInputAndBoundaries(t *testing.T) {
	s, db := fixture(t)
	status(t, request(t, s, "POST", "/api/auth/login", credentials{"admin", testPassword}, nil, map[string]string{"Content-Type": "text/plain"}), 415)
	status(t, request(t, s, "POST", "/api/auth/login", map[string]any{"unknown": true}, nil, nil), 400)
	status(t, request(t, s, "POST", "/api/auth/login", map[string]string{"username": strings.Repeat("x", 20*1024)}, nil, nil), 413)
	for _, body := range []string{`null`, `{"username":"x"} {}`, `{`} {
		r := httptest.NewRequest("POST", "http://example.com/api/auth/login", strings.NewReader(body))
		r.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		s.Handler().ServeHTTP(w, r)
		status(t, w, 400)
	}
	status(t, request(t, s, "POST", "/api/auth/login", credentials{"admin", testPassword}, nil, map[string]string{"Sec-Fetch-Site": "cross-site"}), 403)
	setupAdmin(t, s)
	s.config.CookieSecure = true
	s.config.PublicURL = "https://example.com"
	admin := loginAs(t, s, "admin", testPassword)
	if !admin.cookie.Secure {
		t.Fatal("HTTPS cookie lacks Secure")
	}
	status(t, request(t, s, "GET", "/does-not-exist", nil, nil, nil), 404)
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	status(t, request(t, s, "GET", "/health/live", nil, nil, nil), 200)
	status(t, request(t, s, "GET", "/health/ready", nil, nil, nil), 503)
}
func TestLoginRateLimit(t *testing.T) {
	s, _ := fixture(t)
	// A peer cannot bypass IP limiting using forwarded headers.
	for i := 0; i < 20; i++ {
		w := request(t, s, "POST", "/api/auth/login", nil, nil, map[string]string{"X-Forwarded-For": fmt.Sprintf("10.0.0.%d", i)})
		status(t, w, 415)
	}
	w := request(t, s, "POST", "/api/auth/login", nil, nil, nil)
	status(t, w, 429)
	if w.Header().Get("Retry-After") == "" {
		t.Fatal("missing Retry-After")
	}
}
func TestRecoveryAndTimeout(t *testing.T) {
	s, _ := fixture(t)
	w := httptest.NewRecorder()
	s.observe(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { panic("test") })).ServeHTTP(w, httptest.NewRequest("GET", "/", nil))
	status(t, w, 500)
	if w.Header().Get("X-Request-ID") == "" {
		t.Fatal("missing request ID")
	}
	done := make(chan struct{})
	slow := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { defer close(done); <-r.Context().Done() })
	w = httptest.NewRecorder()
	s.observe(http.TimeoutHandler(slow, time.Millisecond, `{"error":{"code":"request_timeout"}}`)).ServeHTTP(w, httptest.NewRequest("GET", "/", nil))
	status(t, w, 503)
	<-done
}
