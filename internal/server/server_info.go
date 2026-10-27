package server

import "net/http"

func (s *Server) serverInfo(w http.ResponseWriter, r *http.Request) {
	name := s.config.ServerName
	if name == "" {
		name = "Dinosaur"
	}
	writeJSON(w, 200, map[string]any{
		"id": s.serverID, "name": name, "api_version": 1,
		"capabilities": []string{"mobile_qr_login", "tv_pairing", "device_sessions"},
	})
}
