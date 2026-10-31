package bangumi

import (
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	"image"
	"image/jpeg"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPosterIsBoundedAndRestrictedToBangumiCovers(t *testing.T) {
	c := New("movie")
	calls := 0
	var data bytes.Buffer
	if err := jpeg.Encode(&data, image.NewRGBA(image.Rect(0, 0, 12, 20)), nil); err != nil {
		t.Fatal(err)
	}
	c.HTTP.Transport = transport(func(r *http.Request) (*http.Response, error) {
		calls++
		if r.URL.Host != "lain.bgm.tv" || r.Header.Get("Authorization") != "" {
			t.Fatal("unsafe image request")
		}
		rpl := reply("")
		rpl.Body = io.NopCloser(bytes.NewReader(data.Bytes()))
		return rpl, nil
	})
	dir := t.TempDir()
	key, err := c.CachePoster(context.Background(), "https://lain.bgm.tv/pic/cover/l/ab/cd/10_cover.jpg", dir)
	if err != nil || key != fmt.Sprintf("%x", sha256.Sum256(data.Bytes())) {
		t.Fatalf("cache: %s %v", key, err)
	}
	if _, err := os.Stat(filepath.Join(dir, "posters", key+".jpg")); err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{"http://127.0.0.1/x.jpg", "https://evil.example/x.jpg", "https://lain.bgm.tv@127.0.0.1/x.jpg", "https://lain.bgm.tv/pic/cover/l/ab/cd/10.jpg?x=1"} {
		if _, err := c.CachePoster(context.Background(), path, dir); err == nil {
			t.Fatalf("unsafe URL accepted: %s", path)
		}
	}
	if calls != 1 {
		t.Fatalf("unsafe URLs fetched: %d", calls)
	}
	c.HTTP.Transport = transport(func(*http.Request) (*http.Response, error) { return reply(strings.Repeat("x", 5*1024*1024+1)), nil })
	if _, err := c.CachePoster(context.Background(), "https://lain.bgm.tv/pic/cover/l/ab/cd/10.jpg", dir); err == nil {
		t.Fatal("oversized image accepted")
	}
	c.HTTP.Transport = transport(func(*http.Request) (*http.Response, error) { return reply("not a jpeg"), nil })
	if _, err := c.CachePoster(context.Background(), "https://lain.bgm.tv/pic/cover/l/ab/cd/10.jpg", dir); err == nil {
		t.Fatal("invalid image accepted")
	}
}
