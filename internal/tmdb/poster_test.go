package tmdb

import (
	"bytes"
	"context"
	"image"
	"image/jpeg"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"testing"
)

func TestPosterCacheConfinesRequestsAndValidatesImages(t *testing.T) {
	var jpegData bytes.Buffer
	if err := jpeg.Encode(&jpegData, image.NewRGBA(image.Rect(0, 0, 10, 15)), nil); err != nil {
		t.Fatal(err)
	}
	client := New()
	calls := 0
	valid := true
	client.HTTP.Transport = roundTrip(func(r *http.Request) (*http.Response, error) {
		calls++
		if r.URL.Host != "image.tmdb.org" || r.URL.Path != "/t/p/w500/abc.jpg" || r.Header.Get("Authorization") != "" || r.URL.RawQuery != "" {
			t.Fatal("unexpected artwork request or credential leak")
		}
		data := jpegData.Bytes()
		if !valid {
			data = []byte("not an image")
		}
		return &http.Response{StatusCode: 200, Body: io.NopCloser(bytes.NewReader(data))}, nil
	})
	root := t.TempDir()
	for _, path := range []string{"https://example.com/abc.jpg", "/../abc.jpg", "//abc.jpg", "/abc.svg"} {
		if _, err := client.CachePoster(context.Background(), path, root); err == nil || calls != 0 {
			t.Fatal("unsafe artwork path accepted")
		}
	}
	key, err := client.CachePoster(context.Background(), "/abc.jpg", root)
	if err != nil || len(key) != 64 {
		t.Fatalf("failed caching valid artwork: %v", err)
	}
	data, err := os.ReadFile(filepath.Join(root, "posters", key+".jpg"))
	if err != nil || !bytes.Equal(data, jpegData.Bytes()) {
		t.Fatal("cached image is incorrect")
	}
	valid = false
	if _, err := client.CachePoster(context.Background(), "/abc.jpg", root); err == nil {
		t.Fatal("invalid JPEG accepted")
	}
}
