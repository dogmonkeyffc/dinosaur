package server

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	"dinosaur/internal/metadata"
)

type metadataTransport func(*http.Request) (*http.Response, error)

func (f metadataTransport) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestMetadataSettingsPermissionsAndSecretLifecycle(t *testing.T) {
	s, db := fixture(t)
	setupAdmin(t, s)
	admin := loginAs(t, s, "admin", testPassword)
	path := "/api/admin/metadata/tmdb"
	update := metadata.SettingsUpdate{CredentialType: "bearer", Language: "zh-CN", Credential: "private-tmdb-test-secret"}
	for _, method := range []string{"GET", "PUT", "DELETE", "POST"} {
		endpoint := path
		if method == "POST" {
			endpoint += "/test"
		}
		status(t, request(t, s, method, endpoint, update, nil, nil), 401)
		if method != "GET" {
			status(t, request(t, s, method, endpoint, update, &admin, map[string]string{"X-CSRF-Token": ""}), 403)
		}
	}
	viewerHash, err := s.hasher.Hash(context.Background(), testPassword)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.CreateUser(context.Background(), "viewer", viewerHash, "user", false); err != nil {
		t.Fatal(err)
	}
	viewer := loginAs(t, s, "viewer", testPassword)
	for _, method := range []string{"GET", "PUT", "DELETE", "POST"} {
		endpoint := path
		if method == "POST" {
			endpoint += "/test"
		}
		status(t, request(t, s, method, endpoint, update, &viewer, nil), 403)
	}
	get := request(t, s, "GET", path, nil, &admin, nil)
	status(t, get, 200)
	if !strings.Contains(get.Body.String(), `"language":"zh-CN"`) || !strings.Contains(get.Body.String(), `"configured":false`) {
		t.Fatal("incorrect defaults")
	}
	for _, method := range []string{"PUT", "GET"} {
		result := request(t, s, method, path, update, &admin, nil)
		status(t, result, 200)
		if strings.Contains(result.Body.String(), update.Credential) || !strings.Contains(result.Body.String(), `"configured":true`) {
			t.Fatal("secret exposed or not saved")
		}
	}
	status(t, request(t, s, "PUT", path, metadata.SettingsUpdate{Language: "zh-TW"}, &admin, nil), 200)
	saved, err := db.MetadataSettings(context.Background())
	if err != nil || saved.Credential != update.Credential || saved.Language != "zh-TW" {
		t.Fatal("blank input did not preserve secret")
	}
	for _, bad := range []metadata.SettingsUpdate{
		{CredentialType: "api_key"}, {CredentialType: "invalid"}, {Language: "invalid"},
		{Credential: "bad\ncredential"}, {Credential: strings.Repeat("a", 4097)},
	} {
		status(t, request(t, s, "PUT", path, bad, &admin, nil), 400)
	}
	status(t, request(t, s, "DELETE", path, nil, &admin, nil), 200)
	saved, err = db.MetadataSettings(context.Background())
	if err != nil || saved.Credential != "" || saved.Language != "zh-TW" {
		t.Fatal("clear did not preserve preferences")
	}
	status(t, request(t, s, "POST", path+"/test", metadata.SettingsUpdate{}, &admin, nil), 400)
}

func TestMetadataConnectionUsesDraftWithoutSaving(t *testing.T) {
	s, db := fixture(t)
	setupAdmin(t, s)
	admin := loginAs(t, s, "admin", testPassword)
	path := "/api/admin/metadata/tmdb"
	upstreamStatus := 200
	seen := ""
	s.tmdb.HTTP.Transport = metadataTransport(func(r *http.Request) (*http.Response, error) {
		seen = r.Header.Get("Authorization")
		return &http.Response{StatusCode: upstreamStatus, Body: io.NopCloser(strings.NewReader(`{"success":true}`)), Header: make(http.Header)}, nil
	})
	update := metadata.SettingsUpdate{Credential: "draft-secret"}
	status(t, request(t, s, "POST", path+"/test", update, &admin, nil), 200)
	if seen != "Bearer draft-secret" {
		t.Fatal("draft credential not used")
	}
	saved, err := db.MetadataSettings(context.Background())
	if err != nil || saved.Credential != "" {
		t.Fatal("test unexpectedly saved draft")
	}
	status(t, request(t, s, "PUT", path, update, &admin, nil), 200)
	for _, test := range []struct{ upstream, want int }{{200, 200}, {401, 422}, {429, 503}, {500, 502}} {
		upstreamStatus = test.upstream
		result := request(t, s, "POST", path+"/test", metadata.SettingsUpdate{}, &admin, nil)
		status(t, result, test.want)
		if strings.Contains(result.Body.String(), update.Credential) {
			t.Fatal("test response leaked secret")
		}
		var payload map[string]any
		if err := json.Unmarshal(result.Body.Bytes(), &payload); err != nil {
			t.Fatal(err)
		}
	}
}
