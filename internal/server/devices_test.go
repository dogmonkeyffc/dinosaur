package server

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"dinosaur/internal/deviceauth"
)

func mobileTicket(t *testing.T, s *Server, who session) string {
	t.Helper()
	w := request(t, s, "POST", "/api/devices/qr", map[string]any{}, &who, nil)
	status(t, w, 201)
	var result struct {
		Payload string `json:"qr_payload"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	u, err := url.Parse(result.Payload)
	if err != nil {
		t.Fatal(err)
	}
	if u.Scheme != "dinosaur" || u.Host != "login" || u.Query().Get("server") != "http://example.com" {
		t.Fatal("invalid login QR")
	}
	return u.Query().Get("ticket")
}

func exchangeMobile(t *testing.T, s *Server, ticket string) deviceauth.Session {
	t.Helper()
	w := request(t, s, "POST", "/api/devices/qr/exchange", map[string]string{"ticket": ticket, "name": "我的手机"}, nil, nil)
	status(t, w, 200)
	var result deviceauth.Session
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	if len(result.Token) != 43 || len(w.Result().Cookies()) != 0 {
		t.Fatal("invalid native credentials")
	}
	return result
}

func bearer(token string) map[string]string {
	return map[string]string{"Authorization": "Bearer " + token}
}

func TestMobileQRLoginAndAuthenticationSeparation(t *testing.T) {
	s, _ := fixture(t)
	setupAdmin(t, s)
	who := loginAs(t, s, "admin", testPassword)
	ticket := mobileTicket(t, s, who)
	native := exchangeMobile(t, s, ticket)
	status(t, request(t, s, "GET", "/api/auth/me", nil, nil, bearer(native.Token)), 200)
	status(t, request(t, s, "POST", "/api/devices/qr/exchange", map[string]string{"ticket": ticket, "name": "重放"}, nil, nil), 410)
	status(t, request(t, s, "GET", "/api/auth/me", nil, nil, bearer(who.cookie.Value)), 401)
	forged := session{cookie: &http.Cookie{Name: cookieName, Value: native.Token}}
	status(t, request(t, s, "GET", "/api/auth/me", nil, &forged, nil), 401)
	status(t, request(t, s, "GET", "/api/auth/me", nil, &who, bearer(native.Token)), 401)
	status(t, request(t, s, "POST", "/api/devices/qr", map[string]any{}, nil, nil), 401)
	noCSRF := session{cookie: who.cookie}
	status(t, request(t, s, "POST", "/api/devices/qr", map[string]any{}, &noCSRF, nil), 403)
	status(t, request(t, s, "POST", "/api/auth/logout", nil, nil, bearer(native.Token)), 200)
	status(t, request(t, s, "GET", "/api/auth/me", nil, nil, bearer(native.Token)), 401)
}

func TestQRReplacementExpiryAndOwnerRevocation(t *testing.T) {
	s, db := fixture(t)
	setupAdmin(t, s)
	who := loginAs(t, s, "admin", testPassword)
	old := mobileTicket(t, s, who)
	current := mobileTicket(t, s, who)
	status(t, request(t, s, "POST", "/api/devices/qr/exchange", map[string]string{"ticket": old, "name": "手机"}, nil, nil), 410)
	if _, err := db.DB.Exec(`UPDATE device_grants SET expires_at=1`); err != nil {
		t.Fatal(err)
	}
	status(t, request(t, s, "POST", "/api/devices/qr/exchange", map[string]string{"ticket": current, "name": "手机"}, nil, nil), 410)
	last := mobileTicket(t, s, who)
	status(t, request(t, s, "POST", "/api/auth/logout", nil, &who, nil), 200)
	status(t, request(t, s, "POST", "/api/devices/qr/exchange", map[string]string{"ticket": last, "name": "手机"}, nil, nil), 410)
}

func TestDeviceListingAndRevocationOwnership(t *testing.T) {
	s, _ := fixture(t)
	setupAdmin(t, s)
	admin := loginAs(t, s, "admin", testPassword)
	status(t, request(t, s, "POST", "/api/admin/users", credentials{"viewer", testPassword}, &admin, nil), 201)
	viewer := loginAs(t, s, "viewer", testPassword)
	phone := exchangeMobile(t, s, mobileTicket(t, s, viewer))
	w := request(t, s, "GET", "/api/devices", nil, nil, bearer(phone.Token))
	status(t, w, 200)
	var result struct {
		Devices []deviceauth.Device `json:"devices"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	if len(result.Devices) != 1 || !result.Devices[0].Current || strings.Contains(w.Body.String(), phone.Token) || strings.Contains(w.Body.String(), "token_hash") {
		t.Fatal("unsafe device response")
	}
	id := result.Devices[0].ID
	status(t, request(t, s, "DELETE", "/api/devices/"+id, nil, &admin, nil), 200)
	status(t, request(t, s, "GET", "/api/auth/me", nil, nil, bearer(phone.Token)), 200)
	status(t, request(t, s, "DELETE", "/api/devices/"+id, nil, &viewer, nil), 200)
	status(t, request(t, s, "GET", "/api/auth/me", nil, nil, bearer(phone.Token)), 401)
}
