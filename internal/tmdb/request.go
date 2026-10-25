package tmdb

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"

	"dinosaur/internal/metadata"
)

func (c *Client) getJSON(ctx context.Context, settings metadata.Settings, path string, query url.Values, result any) error {
	if settings.Credential == "" {
		return metadata.ErrMissingCredential
	}
	if query == nil {
		query = make(url.Values)
	}
	query.Set("language", settings.Language)
	if settings.CredentialType == "api_key" {
		query.Set("api_key", settings.Credential)
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.themoviedb.org/3"+path+"?"+query.Encode(), nil)
	if err != nil {
		return ErrUnavailable
	}
	request.Header.Set("Accept", "application/json")
	if settings.CredentialType == "bearer" {
		request.Header.Set("Authorization", "Bearer "+settings.Credential)
	}
	response, err := c.HTTP.Do(request)
	if err != nil {
		return ErrUnavailable
	}
	defer response.Body.Close()
	switch response.StatusCode {
	case http.StatusUnauthorized, http.StatusForbidden:
		return ErrUnauthorized
	case http.StatusTooManyRequests:
		return ErrRateLimited
	case http.StatusNotFound:
		return metadata.ErrProviderNotFound
	case http.StatusOK:
		data, err := io.ReadAll(io.LimitReader(response.Body, 2*1024*1024+1))
		if err == nil && len(data) <= 2*1024*1024 && json.Unmarshal(data, result) == nil {
			return nil
		}
	}
	return ErrUnavailable
}
