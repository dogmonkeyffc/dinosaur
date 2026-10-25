package server

import (
	"bytes"
	"fmt"
	"io/fs"
	"net/http"
	"path"
	"strings"
	"time"

	"dinosaur/web"
)

func (s *Server) mountWeb(mux *http.ServeMux) error {
	assets, err := fs.Sub(web.Files, "dist")
	if err != nil {
		return err
	}
	page, err := fs.ReadFile(assets, "index.html")
	if err != nil {
		return fmt.Errorf("React assets are missing; run ./build.sh before starting Go: %w", err)
	}
	render := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("Cache-Control", "no-store")
		// Artplayer creates a version link with the exact style "width:100%;".
		// Permit only that attribute hash, keeping scripts and style elements local.
		w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self'; style-src 'self'; style-src-attr 'unsafe-hashes' 'sha256-nMxMqdZhkHxz5vAuW/PAoLvECzzsmeAxD/BNwG15HuA='; img-src 'self' data:; connect-src 'self'; object-src 'none'; base-uri 'none'; frame-ancestors 'none'; form-action 'self'")
		w.Header().Set("Referrer-Policy", "same-origin")
		_, _ = w.Write(page)
	}
	for _, path := range []string{"GET /{$}", "GET /login", "GET /setup", "GET /settings", "GET /settings/account", "GET /settings/metadata", "GET /settings/scraping", "GET /settings/system", "GET /movies", "GET /movies/{id}", "GET /series", "GET /series/{id}", "GET /watch/{id}"} {
		mux.HandleFunc(path, render)
	}
	mux.HandleFunc("GET /assets/{asset...}", func(w http.ResponseWriter, r *http.Request) { serveAsset(w, r, assets) })
	return nil
}

func serveAsset(w http.ResponseWriter, r *http.Request, assets fs.FS) {
	name := r.PathValue("asset")
	kinds := map[string]string{
		".css": "text/css; charset=utf-8", ".js": "text/javascript; charset=utf-8",
		".svg": "image/svg+xml", ".txt": "text/plain; charset=utf-8", ".json": "application/json",
	}
	kind := kinds[path.Ext(name)]
	if kind == "" || !fs.ValidPath(name) {
		fail(w, 404, "not_found", "asset not found")
		return
	}
	data, err := fs.ReadFile(assets, "assets/"+name)
	if err != nil {
		fail(w, 404, "not_found", "asset not found")
		return
	}
	w.Header().Set("Content-Type", kind)
	if strings.HasPrefix(name, "vendor/") {
		w.Header().Set("Cache-Control", "no-cache")
	} else {
		w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
	}
	http.ServeContent(w, r, name, time.Time{}, bytes.NewReader(data))
}
