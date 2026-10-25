package tmdb

import (
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	"image/jpeg"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
)

var posterPath = regexp.MustCompile(`^/[a-zA-Z0-9]+\.jpg$`)

// CachePoster fetches only TMDB JPEG artwork, never a provider-supplied URL.
func (c *Client) CachePoster(ctx context.Context, path, dataDir string) (string, error) {
	if !posterPath.MatchString(path) {
		return "", ErrUnavailable
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://image.tmdb.org/t/p/w500"+path, nil)
	if err != nil {
		return "", ErrUnavailable
	}
	response, err := c.HTTP.Do(request)
	if err != nil {
		return "", ErrUnavailable
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return "", ErrUnavailable
	}
	data, err := io.ReadAll(io.LimitReader(response.Body, 5*1024*1024+1))
	if err != nil || len(data) > 5*1024*1024 {
		return "", ErrUnavailable
	}
	config, err := jpeg.DecodeConfig(bytes.NewReader(data))
	if err != nil || config.Width <= 0 || config.Width > 2000 || config.Height <= 0 || config.Height > 4000 {
		return "", ErrUnavailable
	}
	if _, err := jpeg.Decode(bytes.NewReader(data)); err != nil {
		return "", ErrUnavailable
	}
	key := fmt.Sprintf("%x", sha256.Sum256(data))
	directory := filepath.Join(dataDir, "posters")
	if err := os.MkdirAll(directory, 0700); err != nil {
		return "", err
	}
	file, err := os.CreateTemp(directory, ".tmdb-*.jpg")
	if err != nil {
		return "", err
	}
	defer os.Remove(file.Name())
	_, writeErr := file.Write(data)
	closeErr := file.Close()
	if writeErr != nil {
		return "", writeErr
	}
	if closeErr != nil {
		return "", closeErr
	}
	return key, os.Rename(file.Name(), filepath.Join(directory, key+".jpg"))
}
