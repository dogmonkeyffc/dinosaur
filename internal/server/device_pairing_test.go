package server

import (
	"encoding/json"
	"testing"

	"dinosaur/internal/deviceauth"
)

type tvChallenge struct {
	ID     string `json:"pairing_id"`
	Code   string `json:"code"`
	Secret string `json:"device_code"`
}

func createTV(t *testing.T, s *Server) tvChallenge {
	t.Helper()
	w := request(t, s, "POST", "/api/devices/pairings", map[string]string{"name": "客厅电视"}, nil, nil)
	status(t, w, 201)
	var result tvChallenge
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	if !deviceauth.ValidCode(result.Code) || len(result.Secret) != 43 {
		t.Fatal("invalid challenge")
	}
	return result
}

func TestTVLoginUsesPhoneSelectedAccount(t *testing.T) {
	s, _ := fixture(t)
	setupAdmin(t, s)
	admin := loginAs(t, s, "admin", testPassword)
	status(t, request(t, s, "POST", "/api/admin/users", credentials{"viewer", testPassword}, &admin, nil), 201)
	viewer := loginAs(t, s, "viewer", testPassword)
	phone := exchangeMobile(t, s, mobileTicket(t, s, viewer))
	tv := createTV(t, s)
	poll := map[string]string{"device_code": tv.Secret}
	status(t, request(t, s, "POST", "/api/devices/pairings/poll", poll, nil, nil), 202)
	w := request(t, s, "POST", "/api/devices/pairings/lookup", map[string]string{"code": tv.Code}, nil, bearer(phone.Token))
	status(t, w, 200)
	var found deviceauth.Grant
	if err := json.Unmarshal(w.Body.Bytes(), &found); err != nil {
		t.Fatal(err)
	}
	if found.ID != tv.ID || found.Name != "客厅电视" {
		t.Fatal("wrong TV pairing")
	}
	approve := map[string]string{"pairing_id": tv.ID}
	status(t, request(t, s, "POST", "/api/devices/pairings/approve", approve, nil, nil), 401)
	status(t, request(t, s, "POST", "/api/devices/pairings/approve", approve, nil, bearer(phone.Token)), 200)
	status(t, request(t, s, "POST", "/api/devices/pairings/approve", approve, &admin, nil), 410)
	w = request(t, s, "POST", "/api/devices/pairings/poll", poll, nil, nil)
	status(t, w, 200)
	var session deviceauth.Session
	if err := json.Unmarshal(w.Body.Bytes(), &session); err != nil {
		t.Fatal(err)
	}
	if session.User.ID != phone.User.ID || session.User.Username != "viewer" || session.User.Role != "user" {
		t.Fatal("TV got a different account or elevated role")
	}
	status(t, request(t, s, "GET", "/api/auth/me", nil, nil, bearer(session.Token)), 200)
	status(t, request(t, s, "GET", "/api/admin/users", nil, nil, bearer(session.Token)), 403)
	status(t, request(t, s, "POST", "/api/devices/pairings/poll", poll, nil, nil), 410)
}

func TestPairingApprovalDiesWithApproverSession(t *testing.T) {
	s, _ := fixture(t)
	setupAdmin(t, s)
	admin := loginAs(t, s, "admin", testPassword)
	tv := createTV(t, s)
	status(t, request(t, s, "POST", "/api/devices/pairings/approve", map[string]string{"pairing_id": tv.ID}, &admin, nil), 200)
	status(t, request(t, s, "POST", "/api/auth/logout", nil, &admin, nil), 200)
	status(t, request(t, s, "POST", "/api/devices/pairings/poll", map[string]string{"device_code": tv.Secret}, nil, nil), 410)
}

func TestPairingValidationAndGuessRate(t *testing.T) {
	s, _ := fixture(t)
	setupAdmin(t, s)
	admin := loginAs(t, s, "admin", testPassword)
	status(t, request(t, s, "POST", "/api/devices/pairings", map[string]string{"name": "\n"}, nil, nil), 400)
	status(t, request(t, s, "POST", "/api/devices/pairings/poll", map[string]string{"device_code": "123456"}, nil, nil), 400)
	for i := 0; i < 10; i++ {
		status(t, request(t, s, "POST", "/api/devices/pairings/lookup", map[string]string{"code": "123456"}, &admin, nil), 410)
	}
	status(t, request(t, s, "POST", "/api/devices/pairings/lookup", map[string]string{"code": "123456"}, &admin, nil), 429)
}
