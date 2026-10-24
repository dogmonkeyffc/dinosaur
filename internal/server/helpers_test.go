package server

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"dinosaur/internal/config"
	"dinosaur/internal/store"
)

const setupSecret = "test-setup-token-with-at-least-32-bytes"
const testPassword = "correct horse battery staple"

type session struct {
	cookie *http.Cookie
	csrf   string
}

func fixture(t *testing.T) (*Server, *store.Store) {
	t.Helper()
	db, err := store.Open(context.Background(), filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	c := config.Config{SetupToken: setupSecret, SessionTTL: time.Hour, RequestTimeout: 5 * time.Second, PublicURL: "http://example.com", DataDir: t.TempDir()}
	s, err := New(c, db, slog.New(slog.NewTextHandler(io.Discard, nil)))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(s.Close)
	return s, db
}
func request(t *testing.T, s *Server, method, path string, body any, who *session, headers map[string]string) *httptest.ResponseRecorder {
	t.Helper()
	var data string
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			t.Fatal(err)
		}
		data = string(b)
	}
	r := httptest.NewRequest(method, "http://example.com"+path, strings.NewReader(data))
	if body != nil {
		r.Header.Set("Content-Type", "application/json")
	}
	if who != nil {
		r.AddCookie(who.cookie)
		r.Header.Set("X-CSRF-Token", who.csrf)
	}
	for k, v := range headers {
		r.Header.Set(k, v)
	}
	w := httptest.NewRecorder()
	s.Handler().ServeHTTP(w, r)
	return w
}
func status(t *testing.T, w *httptest.ResponseRecorder, want int) {
	t.Helper()
	if w.Code != want {
		t.Fatalf("status=%d want=%d body=%s", w.Code, want, w.Body.String())
	}
	if !strings.HasPrefix(w.Header().Get("Content-Type"), "application/json") {
		t.Fatalf("not JSON: %v", w.Header())
	}
	if strings.Contains(w.Body.String(), "password_hash") || strings.Contains(w.Body.String(), "$argon2id$") {
		t.Fatal("password hash leaked")
	}
}
func setupAdmin(t *testing.T, s *Server) {
	t.Helper()
	status(t, request(t, s, "POST", "/api/setup", credentials{"admin", testPassword}, nil, map[string]string{"X-Setup-Token": setupSecret}), 201)
}
func loginAs(t *testing.T, s *Server, name, password string) session {
	t.Helper()
	w := request(t, s, "POST", "/api/auth/login", credentials{name, password}, nil, nil)
	status(t, w, 200)
	var body struct {
		CSRF string `json:"csrf_token"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	cookies := w.Result().Cookies()
	if len(cookies) != 1 || body.CSRF == "" {
		t.Fatal("missing login credentials")
	}
	c := cookies[0]
	if !c.HttpOnly || c.Path != "/" || c.SameSite != http.SameSiteLaxMode || c.MaxAge != 3600 {
		t.Fatalf("bad cookie: %v", c)
	}
	return session{c, body.CSRF}
}
