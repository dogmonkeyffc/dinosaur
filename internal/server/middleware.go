package server

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"strings"
	"time"

	"dinosaur/internal/auth"
)

type responseWriter struct {
	http.ResponseWriter
	status int
}

func (w *responseWriter) WriteHeader(status int) {
	if w.status != 0 {
		return
	}
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}
func (w *responseWriter) Write(b []byte) (int, error) {
	if w.status == 0 {
		w.WriteHeader(200)
	}
	return w.ResponseWriter.Write(b)
}
func (w *responseWriter) Unwrap() http.ResponseWriter { return w.ResponseWriter }
func (s *Server) observe(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		started := time.Now()
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		id, err := auth.NewToken()
		if err != nil {
			s.internal(w, r, err)
			return
		}
		w.Header().Set("X-Request-ID", id)
		rw := &responseWriter{ResponseWriter: w}
		defer func() {
			if recovered := recover(); recovered != nil {
				s.logger.Error("request panic", "request_id", id, "method", r.Method, "path", r.URL.Path)
				if rw.status == 0 {
					fail(rw, 500, "internal_error", "internal server error")
				}
			}
			status := rw.status
			if status == 0 {
				status = 200
			}
			s.logger.Info("http request", "request_id", id, "method", r.Method, "path", r.URL.Path, "status", status, "duration_ms", time.Since(started).Milliseconds())
		}()
		next.ServeHTTP(rw, r)
	})
}
func safeMethod(method string) bool {
	return method == http.MethodGet || method == http.MethodHead || method == http.MethodOptions
}
func (s *Server) checkOrigin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !safeMethod(r.Method) {
			expected := s.config.PublicURL
			if expected == "" {
				scheme := "http"
				if s.config.CookieSecure {
					scheme = "https"
				}
				expected = scheme + "://" + r.Host
			}
			origin := r.Header.Get("Origin")
			if (origin != "" && origin != expected) || r.Header.Get("Sec-Fetch-Site") == "cross-site" {
				fail(w, 403, "invalid_origin", "cross-origin write requests are not allowed")
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}
func (s *Server) require(admin bool, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, cookieErr := r.Cookie(cookieName)
		token, kind := "", "browser"
		authorization := r.Header.Get("Authorization")
		if authorization != "" {
			if cookieErr == nil || len(r.Header.Values("Authorization")) != 1 || !strings.HasPrefix(authorization, "Bearer ") {
				fail(w, 401, "unauthorized", "ambiguous or invalid authentication")
				return
			}
			token, kind = strings.TrimPrefix(authorization, "Bearer "), "device"
		} else if cookieErr == nil {
			token = cookie.Value
		}
		if len(token) != 43 {
			fail(w, 401, "unauthorized", "authentication required")
			return
		}
		u, err := s.store.SessionUserKind(r.Context(), auth.TokenHash(token), kind)
		if errors.Is(err, sql.ErrNoRows) {
			if kind == "browser" {
				s.clearCookie(w)
			}
			fail(w, 401, "unauthorized", "session expired or revoked")
			return
		}
		if err != nil {
			s.internal(w, r, err)
			return
		}
		if admin && u.Role != "admin" {
			fail(w, 403, "forbidden", "administrator permission required")
			return
		}
		if kind == "browser" && !safeMethod(r.Method) && !auth.Equal(r.Header.Get("X-CSRF-Token"), auth.CSRFToken(token)) {
			fail(w, 403, "invalid_csrf", "valid X-CSRF-Token header required")
			return
		}
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), identityKey{}, identity{user: u, token: token, device: kind == "device"})))
	})
}
func current(r *http.Request) identity { return r.Context().Value(identityKey{}).(identity) }
