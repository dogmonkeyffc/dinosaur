package server

import (
	"net/http"

	"dinosaur/internal/auth"
)

func browserOnly(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if current(r).device {
			fail(w, 403, "browser_session_required", "browser session required")
			return
		}
		next(w, r)
	}
}

func (s *Server) createLoginCode(w http.ResponseWriter, r *http.Request) {
	if !s.deviceUserRate(w, r, "login-code", 20) {
		return
	}
	var body struct{}
	if !decodeJSON(w, r, &body) {
		return
	}
	who := current(r)
	grant, code, err := s.devices.CreateLoginCode(r.Context(), auth.TokenHash(who.token))
	if s.deviceError(w, r, err) {
		return
	}
	writeJSON(w, 201, map[string]any{"code": code, "expires_at": grant.ExpiresAt, "user": who.user, "server_url": s.deviceOrigin(r)})
}

func (s *Server) exchangeLoginCode(w http.ResponseWriter, r *http.Request) {
	// Bound aggregate guessing across distributed peers as well as each peer.
	// Charge all attempts (including malformed and successful ones) conservatively.
	if !s.rate(w, auth.Limit{Key: "login-code-global", Max: 100}, auth.Limit{Key: "login-code-ip:" + peerIP(r), Max: 10}) {
		return
	}
	var body struct {
		Code string `json:"code"`
		Name string `json:"name"`
	}
	if !decodeJSON(w, r, &body) {
		return
	}
	session, err := s.devices.RedeemLoginCode(r.Context(), body.Code, body.Name)
	if s.deviceError(w, r, err) {
		return
	}
	writeJSON(w, 200, session)
}

func (s *Server) deviceOrigin(r *http.Request) string {
	if s.config.PublicURL != "" {
		return s.config.PublicURL
	}
	scheme := "http"
	if s.config.CookieSecure || r.TLS != nil {
		scheme = "https"
	}
	return scheme + "://" + r.Host
}
