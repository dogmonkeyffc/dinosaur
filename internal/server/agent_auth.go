package server

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"dinosaur/internal/agent"
	"dinosaur/internal/auth"
)

const agentTokenPrefix = "dino_agent_"

type agentIdentityKey struct{}

func agentKey(r *http.Request) agent.Key { return r.Context().Value(agentIdentityKey{}).(agent.Key) }

func (s *Server) requireAgent(write bool, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		header := r.Header.Get("Authorization")
		token := strings.TrimPrefix(header, "Bearer ")
		_, cookieErr := r.Cookie(cookieName)
		if cookieErr == nil || len(r.Header.Values("Authorization")) != 1 || !strings.HasPrefix(header, "Bearer "+agentTokenPrefix) || len(token) != len(agentTokenPrefix)+43 {
			fail(w, 401, "invalid_agent_key", "a dedicated agent API key is required")
			return
		}
		key, err := s.store.AuthenticateAgentKey(r.Context(), auth.TokenHash(token))
		if errors.Is(err, sql.ErrNoRows) {
			fail(w, 401, "invalid_agent_key", "agent key is invalid, expired or revoked")
			return
		}
		if err != nil {
			s.internal(w, r, err)
			return
		}
		if write && key.Access != "read_write" {
			fail(w, 403, "agent_forbidden", "agent key is read-only")
			return
		}
		if !s.rate(w, auth.Limit{Key: "agent:" + strconv.FormatInt(key.ID, 10), Max: 1000}) {
			return
		}
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), agentIdentityKey{}, key)))
	})
}

func (s *Server) listAgentKeys(w http.ResponseWriter, r *http.Request) {
	keys, err := s.store.AgentKeys(r.Context(), current(r).user.ID)
	if err != nil {
		s.internal(w, r, err)
		return
	}
	writeJSON(w, 200, map[string]any{"keys": keys})
}

func (s *Server) createAgentKey(w http.ResponseWriter, r *http.Request) {
	var input agent.CreateKey
	if !decodeJSON(w, r, &input) {
		return
	}
	input, err := agent.ValidateKey(input)
	if s.agentError(w, r, err) {
		return
	}
	if input.LibraryID > 0 {
		if _, err := s.store.LibraryByID(r.Context(), input.LibraryID); s.agentError(w, r, err) {
			return
		}
	}
	random, err := auth.NewToken()
	if err != nil {
		s.internal(w, r, err)
		return
	}
	token := agentTokenPrefix + random
	key := agent.NewKey(current(r).user.ID, input, agentTokenPrefix+"…"+random[len(random)-6:])
	key, err = s.store.CreateAgentKey(r.Context(), key, auth.TokenHash(token))
	if s.agentError(w, r, err) {
		return
	}
	writeJSON(w, 201, map[string]any{"key": key, "token": token})
}

func (s *Server) revokeAgentKey(w http.ResponseWriter, r *http.Request) {
	id, ok := libraryID(w, r)
	if !ok {
		return
	}
	if s.agentError(w, r, s.store.RevokeAgentKey(r.Context(), current(r).user.ID, id)) {
		return
	}
	writeJSON(w, 200, map[string]bool{"revoked": true})
}
