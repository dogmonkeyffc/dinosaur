package server

import (
	"database/sql"
	"errors"
	"mime"
	"net/http"
	"time"

	"dinosaur/internal/auth"
)

func (s *Server) createMobileAdminGrant(w http.ResponseWriter, r *http.Request) {
	if !current(r).device {
		fail(w, 403, "forbidden", "device session required")
		return
	}
	var body struct{}
	if !decodeJSON(w, r, &body) {
		return
	}
	if !s.deviceUserRate(w, r, "admin-web", 20) {
		return
	}
	secret, err := auth.NewToken()
	if err != nil {
		s.internal(w, r, err)
		return
	}
	err = s.store.CreateMobileAdminGrant(r.Context(), auth.TokenHash(current(r).token), auth.TokenHash(secret))
	if err != nil {
		s.internal(w, r, err)
		return
	}
	writeJSON(w, 200, map[string]any{"code": secret})
}

func (s *Server) closeMobileAdminSessions(w http.ResponseWriter, r *http.Request) {
	if !current(r).device {
		fail(w, 403, "forbidden", "device session required")
		return
	}
	if err := s.store.CloseMobileAdminSessions(r.Context(), auth.TokenHash(current(r).token)); err != nil {
		s.internal(w, r, err)
		return
	}
	writeJSON(w, 200, map[string]bool{"ok": true})
}

func (s *Server) exchangeMobileAdminGrant(w http.ResponseWriter, r *http.Request) {
	if !s.rate(w, auth.Limit{Key: "admin-web-exchange:" + peerIP(r), Max: 40}) {
		return
	}
	kind, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil || kind != "application/x-www-form-urlencoded" {
		fail(w, 415, "unsupported_media_type", "form body required")
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, 1024)
	if err := r.ParseForm(); err != nil || len(r.PostForm) != 1 || len(r.PostForm["code"]) != 1 || len(r.PostForm.Get("code")) != 43 {
		fail(w, 400, "invalid_pairing", "invalid admin login code")
		return
	}
	token, err := auth.NewToken()
	if err != nil {
		s.internal(w, r, err)
		return
	}
	ttl := min(s.config.SessionTTL, time.Hour)
	expires, err := s.store.RedeemMobileAdminGrant(r.Context(), auth.TokenHash(r.PostForm.Get("code")), auth.TokenHash(token), time.Now().Add(ttl))
	if errors.Is(err, sql.ErrNoRows) {
		fail(w, 401, "unauthorized", "admin login expired or already used")
		return
	}
	if err != nil {
		s.internal(w, r, err)
		return
	}
	s.setCookie(w, token, expires)
	http.Redirect(w, r, "/settings", http.StatusSeeOther)
}
