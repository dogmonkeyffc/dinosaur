package server

import "testing"

func TestTVCancellationRequiresPrivateSecretAndInvalidatesCode(t *testing.T) {
	s, _ := fixture(t)
	setupAdmin(t, s)
	admin := loginAs(t, s, "admin", testPassword)
	tv := createTV(t, s)
	status(t, request(t, s, "POST", "/api/devices/pairings/cancel", map[string]string{"device_code": tv.Code}, nil, nil), 400)
	status(t, request(t, s, "POST", "/api/devices/pairings/poll", map[string]string{"device_code": tv.Secret}, nil, nil), 202)
	status(t, request(t, s, "POST", "/api/devices/pairings/cancel", map[string]string{"device_code": tv.Secret}, nil, nil), 200)
	status(t, request(t, s, "POST", "/api/devices/pairings/cancel", map[string]string{"device_code": tv.Secret}, nil, nil), 200)
	status(t, request(t, s, "POST", "/api/devices/pairings/poll", map[string]string{"device_code": tv.Secret}, nil, nil), 410)
	status(t, request(t, s, "POST", "/api/devices/pairings/lookup", map[string]string{"code": tv.Code}, &admin, nil), 410)
}
