package server

import (
	"net/http"

	"dinosaur/internal/auth"
	"dinosaur/internal/deviceauth"
)

func (s *Server) createTVPairing(w http.ResponseWriter, r *http.Request) {
	if !s.rate(w, auth.Limit{Key: "tv-pairing:" + peerIP(r), Max: 30}) {
		return
	}
	var body struct {
		Name string `json:"name"`
	}
	if !decodeJSON(w, r, &body) {
		return
	}
	grant, secret, err := s.devices.Create(r.Context(), "tv", "", body.Name)
	if s.deviceError(w, r, err) {
		return
	}
	writeJSON(w, 201, map[string]any{"pairing_id": grant.ID, "name": grant.Name, "code": grant.Code, "device_code": secret, "expires_at": grant.ExpiresAt, "interval": 3})
}

func (s *Server) cancelTVPairing(w http.ResponseWriter, r *http.Request) {
	if !s.rate(w, auth.Limit{Key: "tv-cancel:" + peerIP(r), Max: 60}) {
		return
	}
	var body struct {
		DeviceCode string `json:"device_code"`
	}
	if !decodeJSON(w, r, &body) {
		return
	}
	if s.deviceError(w, r, s.devices.Cancel(r.Context(), body.DeviceCode)) {
		return
	}
	writeJSON(w, 200, map[string]bool{"ok": true})
}

func (s *Server) pollTVPairing(w http.ResponseWriter, r *http.Request) {
	var body struct {
		DeviceCode string `json:"device_code"`
	}
	if !decodeJSON(w, r, &body) {
		return
	}
	if len(body.DeviceCode) != 43 {
		s.deviceError(w, r, deviceauth.ErrInvalid)
		return
	}
	if !s.rate(w,
		auth.Limit{Key: "tv-poll-ip:" + peerIP(r), Max: 1200},
		auth.Limit{Key: "tv-poll:" + auth.TokenHash(body.DeviceCode), Max: 120},
	) {
		return
	}
	session, err := s.devices.Redeem(r.Context(), "tv", body.DeviceCode, "")
	if s.deviceError(w, r, err) {
		return
	}
	writeJSON(w, 200, session)
}

func (s *Server) lookupTVPairing(w http.ResponseWriter, r *http.Request) {
	if !s.deviceUserRate(w, r, "tv-lookup", 10) {
		return
	}
	var body struct {
		Code string `json:"code"`
	}
	if !decodeJSON(w, r, &body) {
		return
	}
	grant, err := s.devices.Lookup(r.Context(), body.Code)
	if s.deviceError(w, r, err) {
		return
	}
	writeJSON(w, 200, grant)
}

func (s *Server) approveTVPairing(w http.ResponseWriter, r *http.Request) {
	if !s.deviceUserRate(w, r, "tv-approve", 20) {
		return
	}
	var body struct {
		ID string `json:"pairing_id"`
	}
	if !decodeJSON(w, r, &body) {
		return
	}
	if len(body.ID) != 43 {
		s.deviceError(w, r, deviceauth.ErrInvalid)
		return
	}
	err := s.devices.Approve(r.Context(), body.ID, auth.TokenHash(current(r).token))
	if s.deviceError(w, r, err) {
		return
	}
	writeJSON(w, 200, map[string]any{"ok": true, "user": current(r).user})
}
