package tmdb

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	"dinosaur/internal/metadata"
)

type roundTrip func(*http.Request) (*http.Response, error)

func (f roundTrip) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestCredentialAuthenticationAndFailures(t *testing.T) {
	for _, kind := range []string{"bearer", "api_key"} {
		for _, test := range []struct {
			status int
			body   string
			want   error
		}{
			{200, `{"success":true}`, nil},
			{200, `{"success":false}`, ErrUnavailable},
			{200, `not JSON`, ErrUnavailable},
			{401, `secret echoed upstream`, ErrUnauthorized},
			{403, ``, ErrUnauthorized},
			{429, ``, ErrRateLimited},
			{500, ``, ErrUnavailable},
			{302, ``, ErrUnavailable},
		} {
			client := New()
			client.HTTP.Transport = roundTrip(func(r *http.Request) (*http.Response, error) {
				if r.URL.Scheme != "https" || r.URL.Host != "api.themoviedb.org" || r.URL.Path != "/3/authentication" {
					t.Fatal("unexpected upstream endpoint")
				}
				if kind == "bearer" && (r.Header.Get("Authorization") != "Bearer private-test-secret" || r.URL.RawQuery != "") {
					t.Fatal("invalid bearer authentication")
				}
				if kind == "api_key" && (r.URL.Query().Get("api_key") != "private-test-secret" || r.Header.Get("Authorization") != "") {
					t.Fatal("invalid key authentication")
				}
				return &http.Response{StatusCode: test.status, Body: io.NopCloser(strings.NewReader(test.body)), Header: make(http.Header)}, nil
			})
			settings := metadata.Settings{CredentialType: kind, Language: "zh-CN", Credential: "private-test-secret"}
			if err := client.Test(context.Background(), settings); !errors.Is(err, test.want) {
				t.Fatalf("kind=%s status=%d err=%v want=%v", kind, test.status, err, test.want)
			}
		}
	}
}

func TestTransportErrorRedactionAndMissingCredential(t *testing.T) {
	client := New()
	called := false
	client.HTTP.Transport = roundTrip(func(r *http.Request) (*http.Response, error) {
		called = true
		return nil, errors.New("request failed for " + r.URL.String())
	})
	if err := client.Test(context.Background(), metadata.DefaultSettings()); !errors.Is(err, metadata.ErrMissingCredential) || called {
		t.Fatal("missing credential must not send a request")
	}
	settings := metadata.Settings{CredentialType: "api_key", Credential: "private-test-secret"}
	err := client.Test(context.Background(), settings)
	if !errors.Is(err, ErrUnavailable) || strings.Contains(err.Error(), settings.Credential) {
		t.Fatal("transport error leaked credentials")
	}
	if client.HTTP.Timeout == 0 || client.HTTP.CheckRedirect == nil {
		t.Fatal("missing timeout or redirect protection")
	}
}
