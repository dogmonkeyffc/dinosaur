package bangumi

import (
	"bytes"
	"context"
	"crypto/sha256"
	"dinosaur/internal/metadata"
	"fmt"
	"image/jpeg"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
)

var posterPath = regexp.MustCompile(`^https://lain\.bgm\.tv/(?:r/(?:100|200|400|800)/)?pic/cover/[a-z]/[a-f0-9]{2}/[a-f0-9]{2}/[a-zA-Z0-9_]+\.jpg$`)

// CachePoster accepts only Bangumi cover URLs, with redirects disabled by New.
func (c *Client) CachePoster(ctx context.Context, path, dataDir string) (string, error) {
	if !posterPath.MatchString(path) {
		return "", metadata.ErrProviderUnavailable
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, path, nil)
	if err != nil {
		return "", metadata.ErrProviderUnavailable
	}
	request.Header.Set("User-Agent", userAgent)
	response, err := c.HTTP.Do(request)
	if err != nil {
		return "", metadata.ErrProviderUnavailable
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return "", metadata.ErrProviderUnavailable
	}
	data, err := io.ReadAll(io.LimitReader(response.Body, 5*1024*1024+1))
	if err != nil || len(data) > 5*1024*1024 {
		return "", metadata.ErrProviderUnavailable
	}
	config, err := jpeg.DecodeConfig(bytes.NewReader(data))
	if err != nil || config.Width <= 0 || config.Width > 2000 || config.Height <= 0 || config.Height > 4000 {
		return "", metadata.ErrProviderUnavailable
	}
	if _, err := jpeg.Decode(bytes.NewReader(data)); err != nil {
		return "", metadata.ErrProviderUnavailable
	}
	key := fmt.Sprintf("%x", sha256.Sum256(data))
	directory := filepath.Join(dataDir, "posters")
	if err := os.MkdirAll(directory, 0700); err != nil {
		return "", err
	}
	file, err := os.CreateTemp(directory, ".bangumi-*.jpg")
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
