package server

import (
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"
	"testing/fstest"
)

func TestEmbeddedReactRoutesAndAssets(t *testing.T) {
	s, _ := fixture(t)
	assets := map[string]bool{}
	for _, path := range []string{"/", "/login", "/setup", "/settings", "/settings/account", "/settings/system", "/settings/metadata", "/settings/scraping", "/movies", "/movies/1", "/series/1", "/watch/1"} {
		w := request(t, s, "GET", path, nil, nil, nil)
		if w.Code != 200 || !strings.HasPrefix(w.Header().Get("Content-Type"), "text/html") || !strings.Contains(w.Body.String(), `id="root"`) {
			t.Fatalf("bad React route %s: %d", path, w.Code)
		}
		if w.Header().Get("Cache-Control") != "no-store" || !strings.Contains(w.Header().Get("Content-Security-Policy"), "frame-ancestors 'none'") {
			t.Fatal("index cache policy or CSP missing")
		}
		for _, match := range regexp.MustCompile(`(?:src|href)="(/assets/[^" ]+)"`).FindAllStringSubmatch(w.Body.String(), -1) {
			assets[match[1]] = true
		}
	}
	if len(assets) < 4 {
		t.Fatal("production HTML must reference bundled JS, CSS and player assets")
	}
	for path := range assets {
		w := request(t, s, "GET", path, nil, nil, nil)
		if w.Code != 200 || w.Body.Len() == 0 {
			t.Fatalf("missing build asset %s: %d", path, w.Code)
		}
		cache := w.Header().Get("Cache-Control")
		if strings.Contains(path, "/vendor/") && cache != "no-cache" || !strings.Contains(path, "/vendor/") && !strings.Contains(cache, "immutable") {
			t.Fatalf("incorrect cache policy for %s: %s", path, cache)
		}
		head := request(t, s, "HEAD", path, nil, nil, nil)
		if head.Code != 200 || head.Body.Len() != 0 {
			t.Fatalf("bad HEAD response for %s", path)
		}
	}
	for _, path := range []string{"/assets/missing.js", "/assets/index.html", "/assets/vendor/", "/api/missing"} {
		status(t, request(t, s, "GET", path, nil, nil, nil), 404)
	}
}

func TestAssetHandlerRejectsTraversalAndDirectories(t *testing.T) {
	assets := fstest.MapFS{"assets/bundle.js": {Data: []byte("console.log('test')")}}
	for _, name := range []string{"../bundle.js", "a/../../bundle.js", "/bundle.js", "", "bundle.html"} {
		r := httptest.NewRequest("GET", "/assets/test", nil)
		r.SetPathValue("asset", name)
		w := httptest.NewRecorder()
		serveAsset(w, r, assets)
		status(t, w, 404)
	}
}

func TestComponentStylesUseUniqueResponseNonce(t *testing.T) {
	s, _ := fixture(t)
	seen := map[string]bool{}
	pattern := regexp.MustCompile(`<meta name="dinosaur-style-nonce" content="([A-Za-z0-9_-]{43})">`)
	for range 2 {
		w := request(t, s, "GET", "/series", nil, nil, nil)
		match := pattern.FindStringSubmatch(w.Body.String())
		if len(match) != 2 || seen[match[1]] {
			t.Fatal("each page must carry a fresh style nonce")
		}
		seen[match[1]] = true
		policy := w.Header().Get("Content-Security-Policy")
		if !strings.Contains(policy, "style-src 'self' 'nonce-"+match[1]+"'") || strings.Contains(policy, "'unsafe-inline'") {
			t.Fatal("component styles must use the nonce without allowing arbitrary inline styles")
		}
	}
}
