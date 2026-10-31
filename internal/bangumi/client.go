// Package bangumi adapts public Bangumi subject metadata; TMDB secrets never leave TMDB.
package bangumi

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"sync"
	"time"

	"dinosaur/internal/metadata"
)

const userAgent = "dogmonkeyffc/dinosaur/0.1 (https://github.com/dogmonkeyffc/dinosaur)"

type Client struct {
	HTTP  *http.Client
	kind  string
	mu    sync.Mutex
	cache map[int64]cachedSubject
	next  time.Time
}

type cachedSubject struct {
	subject subject
	expires time.Time
}

func New(kind string) *Client {
	return &Client{kind: kind, cache: make(map[int64]cachedSubject), HTTP: &http.Client{
		Timeout:       8 * time.Second,
		CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse },
	}}
}

func (c *Client) request(ctx context.Context, method, path string, body any, target any) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	var data []byte
	var err error
	if body != nil {
		data, err = json.Marshal(body)
		if err != nil {
			return metadata.ErrProviderUnavailable
		}
	}
	c.mu.Lock()
	when := maxTime(time.Now(), c.next)
	c.next = when.Add(300 * time.Millisecond)
	c.mu.Unlock()
	timer := time.NewTimer(time.Until(when))
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
	}
	req, err := http.NewRequestWithContext(ctx, method, "https://api.bgm.tv"+path, bytes.NewReader(data))
	if err != nil {
		return metadata.ErrProviderUnavailable
	}
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	response, err := c.HTTP.Do(req)
	if err != nil {
		return metadata.ErrProviderUnavailable
	}
	defer response.Body.Close()
	switch response.StatusCode {
	case 401, 403:
		return metadata.ErrProviderAuth
	case 404:
		return metadata.ErrProviderNotFound
	case 429:
		return metadata.ErrProviderRateLimited
	case 200:
	default:
		return metadata.ErrProviderUnavailable
	}
	data, err = io.ReadAll(io.LimitReader(response.Body, 2*1024*1024+1))
	if err != nil || len(data) > 2*1024*1024 || json.Unmarshal(data, target) != nil {
		return metadata.ErrProviderUnavailable
	}
	return nil
}

func maxTime(a, b time.Time) time.Time {
	if a.After(b) {
		return a
	}
	return b
}
