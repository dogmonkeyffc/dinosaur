package server

import (
	"errors"
	"net/http"
	"strconv"

	"dinosaur/internal/auth"
	"dinosaur/internal/deviceauth"
)

func (s *Server) mountDevices(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/server", s.serverInfo)
	mux.Handle("POST /api/devices/qr", s.require(false, http.HandlerFunc(s.createMobileQR)))
	mux.HandleFunc("POST /api/devices/qr/exchange", s.exchangeMobileQR)
	mux.HandleFunc("POST /api/devices/pairings", s.createTVPairing)
	mux.HandleFunc("POST /api/devices/pairings/poll", s.pollTVPairing)
	mux.HandleFunc("POST /api/devices/pairings/cancel", s.cancelTVPairing)
	mux.Handle("POST /api/devices/pairings/lookup", s.require(false, http.HandlerFunc(s.lookupTVPairing)))
	mux.Handle("POST /api/devices/pairings/approve", s.require(false, http.HandlerFunc(s.approveTVPairing)))
	mux.Handle("GET /api/devices", s.require(false, http.HandlerFunc(s.listDevices)))
	mux.Handle("DELETE /api/devices/{id}", s.require(false, http.HandlerFunc(s.revokeDevice)))
}

func (s *Server) deviceError(w http.ResponseWriter, r *http.Request, err error) bool {
	if err == nil {
		return false
	}
	switch {
	case errors.Is(err, deviceauth.ErrPending):
		writeJSON(w, 202, map[string]any{"status": "pending", "interval": 3})
	case errors.Is(err, deviceauth.ErrExpired):
		fail(w, 410, "pairing_expired", "pairing expired or already used")
	case errors.Is(err, deviceauth.ErrInvalid):
		fail(w, 400, "invalid_pairing", "invalid device name or pairing code")
	case errors.Is(err, deviceauth.ErrConflict):
		fail(w, 409, "pairing_conflict", "pairing already approved")
	case errors.Is(err, deviceauth.ErrFull):
		fail(w, 429, "rate_limited", "too many pending pairings")
	default:
		s.internal(w, r, err)
	}
	return true
}

func (s *Server) deviceUserRate(w http.ResponseWriter, r *http.Request, operation string, max int) bool {
	return s.rate(w, auth.Limit{Key: operation + ":user:" + strconv.FormatInt(current(r).user.ID, 10), Max: max})
}

func (s *Server) listDevices(w http.ResponseWriter, r *http.Request) {
	who := current(r)
	devices, err := s.devices.List(r.Context(), who.user.ID, auth.TokenHash(who.token))
	if s.deviceError(w, r, err) {
		return
	}
	writeJSON(w, 200, map[string]any{"devices": devices})
}

func (s *Server) revokeDevice(w http.ResponseWriter, r *http.Request) {
	if len(r.PathValue("id")) != 43 {
		fail(w, 400, "invalid_id", "invalid device ID")
		return
	}
	err := s.devices.Revoke(r.Context(), current(r).user.ID, r.PathValue("id"))
	if s.deviceError(w, r, err) {
		return
	}
	writeJSON(w, 200, map[string]bool{"ok": true})
}
