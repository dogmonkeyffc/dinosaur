package server

import (
	"database/sql"
	"errors"
	"net/http"
	"strconv"

	"dinosaur/internal/auth"
	"dinosaur/internal/store"
)

func (s *Server) createUser(w http.ResponseWriter, r *http.Request) {
	if !s.rate(w, auth.Limit{Key: "create-user:" + strconv.FormatInt(current(r).user.ID, 10), Max: 20}) {
		return
	}
	var body struct {
		Username string `json:"username"`
		Password string `json:"password"`
		Role     string `json:"role"`
	}
	if !decodeJSON(w, r, &body) {
		return
	}
	body.Username = auth.NormalizeUsername(body.Username)
	if err := auth.ValidateCredentials(body.Username, body.Password); err != nil {
		fail(w, 400, "invalid_credentials", err.Error())
		return
	}
	if body.Role == "" {
		body.Role = "user"
	}
	if body.Role != "admin" && body.Role != "user" {
		fail(w, 400, "invalid_role", "role must be admin or user")
		return
	}
	hash, err := s.hasher.Hash(r.Context(), body.Password)
	if err != nil {
		s.internal(w, r, err)
		return
	}
	u, err := s.store.CreateUser(r.Context(), body.Username, hash, body.Role, false)
	if errors.Is(err, store.ErrConflict) {
		fail(w, 409, "username_exists", "username already exists")
		return
	}
	if err != nil {
		s.internal(w, r, err)
		return
	}
	writeJSON(w, 201, map[string]any{"user": u})
}
func (s *Server) updateUser(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		fail(w, 400, "invalid_id", "invalid user ID")
		return
	}
	var body struct {
		Disabled *bool `json:"disabled"`
	}
	if !decodeJSON(w, r, &body) {
		return
	}
	if body.Disabled == nil {
		fail(w, 400, "invalid_body", "disabled is required")
		return
	}
	err = s.store.SetDisabled(r.Context(), id, *body.Disabled)
	if errors.Is(err, sql.ErrNoRows) {
		fail(w, 404, "not_found", "user not found")
		return
	}
	if errors.Is(err, store.ErrLastAdmin) {
		fail(w, 409, "last_admin", err.Error())
		return
	}
	if err != nil {
		s.internal(w, r, err)
		return
	}
	writeJSON(w, 200, map[string]bool{"ok": true})
}
func (s *Server) changePassword(w http.ResponseWriter, r *http.Request) {
	who := current(r)
	if !s.rate(w, auth.Limit{Key: "password:" + strconv.FormatInt(who.user.ID, 10), Max: 5}) {
		return
	}
	var body struct {
		CurrentPassword string `json:"current_password"`
		NewPassword     string `json:"new_password"`
	}
	if !decodeJSON(w, r, &body) {
		return
	}
	if err := auth.ValidateCredentials(who.user.Username, body.NewPassword); err != nil {
		fail(w, 400, "invalid_password", err.Error())
		return
	}
	if len(body.CurrentPassword) > 1024 {
		fail(w, 400, "invalid_password", "current password is too long")
		return
	}
	if err := s.hasher.Verify(r.Context(), body.CurrentPassword, who.user.PasswordHash); err != nil {
		if errors.Is(err, auth.ErrPassword) {
			fail(w, 403, "invalid_credentials", "current password is incorrect")
			return
		}
		s.internal(w, r, err)
		return
	}
	hash, err := s.hasher.Hash(r.Context(), body.NewPassword)
	if err != nil {
		s.internal(w, r, err)
		return
	}
	if err := s.store.ChangePassword(r.Context(), who.user.ID, hash); err != nil {
		s.internal(w, r, err)
		return
	}
	s.clearCookie(w)
	writeJSON(w, 200, map[string]bool{"ok": true})
}
