package server

import (
	"database/sql"
	"errors"
	"net/http"
	"time"

	"dinosaur/internal/auth"
	"dinosaur/internal/store"
)

type credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (s *Server) setupStatus(w http.ResponseWriter, r *http.Request) {
	initialized, err := s.store.Initialized(r.Context())
	if err != nil {
		s.internal(w, r, err)
		return
	}
	writeJSON(w, 200, map[string]bool{"initialized": initialized})
}
func (s *Server) setup(w http.ResponseWriter, r *http.Request) {
	initialized, err := s.store.Initialized(r.Context())
	if err != nil {
		s.internal(w, r, err)
		return
	}
	if initialized {
		fail(w, 409, "already_initialized", "administrator already initialized")
		return
	}
	if !s.rate(w, auth.Limit{Key: "setup-ip:" + peerIP(r), Max: 10}) {
		return
	}
	if s.config.SetupToken == "" || !auth.Equal(auth.TokenHash(r.Header.Get("X-Setup-Token")), auth.TokenHash(s.config.SetupToken)) {
		fail(w, 403, "invalid_setup_token", "valid X-Setup-Token header required")
		return
	}
	var body credentials
	if !decodeJSON(w, r, &body) {
		return
	}
	body.Username = auth.NormalizeUsername(body.Username)
	if err := auth.ValidateCredentials(body.Username, body.Password); err != nil {
		fail(w, 400, "invalid_credentials", err.Error())
		return
	}
	hash, err := s.hasher.Hash(r.Context(), body.Password)
	if err != nil {
		s.internal(w, r, err)
		return
	}
	u, err := s.store.CreateUser(r.Context(), body.Username, hash, "admin", true)
	if errors.Is(err, store.ErrInitialized) {
		fail(w, 409, "already_initialized", "administrator already initialized")
		return
	}
	if err != nil {
		s.internal(w, r, err)
		return
	}
	writeJSON(w, 201, map[string]any{"user": u})
}
func (s *Server) login(w http.ResponseWriter, r *http.Request) {
	if !s.rate(w, auth.Limit{Key: "login-ip:" + peerIP(r), Max: 20}) {
		return
	}
	var body credentials
	if !decodeJSON(w, r, &body) {
		return
	}
	body.Username = auth.NormalizeUsername(body.Username)
	if len(body.Username) > 32 || len(body.Password) > 1024 {
		fail(w, 401, "invalid_credentials", "invalid username or password")
		return
	}
	if !s.rate(w, auth.Limit{Key: "login-account:" + auth.TokenHash(body.Username), Max: 10}) {
		return
	}
	u, err := s.store.UserByName(r.Context(), body.Username)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		s.internal(w, r, err)
		return
	}
	hash := s.dummyHash
	if err == nil {
		hash = u.PasswordHash
	}
	verifyErr := s.hasher.Verify(r.Context(), body.Password, hash)
	if errors.Is(err, sql.ErrNoRows) || u.Disabled || errors.Is(verifyErr, auth.ErrPassword) {
		fail(w, 401, "invalid_credentials", "invalid username or password")
		return
	}
	if verifyErr != nil {
		s.internal(w, r, verifyErr)
		return
	}
	token, err := auth.NewToken()
	if err != nil {
		s.internal(w, r, err)
		return
	}
	expires := time.Now().Add(s.config.SessionTTL)
	if err := s.store.CreateSession(r.Context(), auth.TokenHash(token), u.ID, u.PasswordHash, expires); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			fail(w, 401, "invalid_credentials", "invalid username or password")
			return
		}
		s.internal(w, r, err)
		return
	}
	s.setCookie(w, token, expires)
	writeJSON(w, 200, map[string]any{"user": u, "csrf_token": auth.CSRFToken(token), "expires_at": expires.Unix()})
}
func (s *Server) me(w http.ResponseWriter, r *http.Request) {
	who := current(r)
	writeJSON(w, 200, map[string]any{"user": who.user, "csrf_token": auth.CSRFToken(who.token)})
}
func (s *Server) logout(w http.ResponseWriter, r *http.Request) {
	if err := s.store.DeleteSession(r.Context(), auth.TokenHash(current(r).token)); err != nil {
		s.internal(w, r, err)
		return
	}
	s.clearCookie(w)
	writeJSON(w, 200, map[string]bool{"ok": true})
}
