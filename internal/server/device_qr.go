package server

import (
	"net/http"
	"net/url"

	"dinosaur/internal/auth"
)

func (s *Server) createMobileQR(w http.ResponseWriter, r *http.Request) {
	if !s.deviceUserRate(w, r, "mobile-qr", 20) {
		return
	}
	var body struct{}
	if !decodeJSON(w, r, &body) {
		return
	}
	who := current(r)
	grant, secret, err := s.devices.Create(r.Context(), "mobile", auth.TokenHash(who.token), "")
	if s.deviceError(w, r, err) {
		return
	}
	origin := s.config.PublicURL
	if origin == "" {
		scheme := "http"
		if s.config.CookieSecure || r.TLS != nil {
			scheme = "https"
		}
		origin = scheme + "://" + r.Host
	}
	query := url.Values{"server": {origin}, "ticket": {secret}}
	writeJSON(w, 201, map[string]any{"qr_payload": "dinosaur://login?" + query.Encode(), "expires_at": grant.ExpiresAt, "user": who.user})
}

func (s *Server) exchangeMobileQR(w http.ResponseWriter, r *http.Request) {
	if !s.rate(w, auth.Limit{Key: "mobile-exchange:" + peerIP(r), Max: 30}) {
		return
	}
	var body struct {
		Ticket string `json:"ticket"`
		Name   string `json:"name"`
	}
	if !decodeJSON(w, r, &body) {
		return
	}
	session, err := s.devices.Redeem(r.Context(), "mobile", body.Ticket, body.Name)
	if s.deviceError(w, r, err) {
		return
	}
	writeJSON(w, 200, session)
}
