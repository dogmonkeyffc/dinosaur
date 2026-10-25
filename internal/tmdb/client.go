// Package tmdb adapts the official TMDB HTTP API to metadata operations.
package tmdb

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"time"

	"dinosaur/internal/metadata"
)

var ErrUnauthorized = metadata.ErrProviderAuth
var ErrUnavailable = metadata.ErrProviderUnavailable
var ErrRateLimited = metadata.ErrProviderRateLimited

type Client struct {
	HTTP *http.Client
}

func New() *Client {
	return &Client{HTTP: &http.Client{
		Timeout:       8 * time.Second,
		CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse },
	}}
}

func (c *Client) Test(ctx context.Context, settings metadata.Settings) error {
	if settings.Credential == "" {
		return metadata.ErrMissingCredential
	}
	endpoint, _ := url.Parse("https://api.themoviedb.org/3/authentication")
	if settings.CredentialType == "api_key" {
		query := endpoint.Query()
		query.Set("api_key", settings.Credential)
		endpoint.RawQuery = query.Encode()
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return ErrUnavailable
	}
	request.Header.Set("Accept", "application/json")
	if settings.CredentialType == "bearer" {
		request.Header.Set("Authorization", "Bearer "+settings.Credential)
	}
	response, err := c.HTTP.Do(request)
	if err != nil {
		// Transport errors can contain a URL with the API key. Never expose them.
		return ErrUnavailable
	}
	defer response.Body.Close()
	switch response.StatusCode {
	case http.StatusUnauthorized, http.StatusForbidden:
		return ErrUnauthorized
	case http.StatusTooManyRequests:
		return ErrRateLimited
	case http.StatusOK:
		var result struct {
			Success bool `json:"success"`
		}
		if json.NewDecoder(io.LimitReader(response.Body, 64*1024)).Decode(&result) == nil && result.Success {
			return nil
		}
	}
	return ErrUnavailable
}
