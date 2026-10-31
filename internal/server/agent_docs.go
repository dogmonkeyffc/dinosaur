package server

import (
	_ "embed"
	"net/http"
)

//go:embed agent-api.md
var agentAPIDocs string

func (s *Server) agentDocs(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(200)
	_, _ = w.Write([]byte(agentAPIDocs))
}
