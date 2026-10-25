package server

import (
	"html/template"
	"net/http"
	"path/filepath"

	"dinosaur/web"
)

func (s *Server) mountWeb(mux *http.ServeMux) error {
	page, err := template.ParseFS(web.Files, "index.html")
	if err != nil {
		return err
	}
	render := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		// Artplayer creates a version link with the exact style "width:100%;".
		// Permit only that attribute hash, keeping scripts and style elements local.
		w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self'; style-src 'self'; style-src-attr 'unsafe-hashes' 'sha256-nMxMqdZhkHxz5vAuW/PAoLvECzzsmeAxD/BNwG15HuA='; img-src 'self' data:; connect-src 'self'; object-src 'none'; base-uri 'none'; frame-ancestors 'none'; form-action 'self'")
		w.Header().Set("Referrer-Policy", "same-origin")
		if err := page.Execute(w, nil); err != nil {
			s.logger.Error("render web page", "error", err)
		}
	}
	for _, path := range []string{"GET /{$}", "GET /login", "GET /setup", "GET /settings", "GET /settings/metadata", "GET /settings/scraping", "GET /movies", "GET /movies/{id}", "GET /series", "GET /series/{id}", "GET /watch/{id}"} {
		mux.HandleFunc(path, render)
	}
	mux.HandleFunc("GET /assets/{name}", func(w http.ResponseWriter, r *http.Request) { serveAsset(w, r, "") })
	mux.HandleFunc("GET /assets/vendor/{name}", func(w http.ResponseWriter, r *http.Request) { serveAsset(w, r, "vendor/") })
	return nil
}

func serveAsset(w http.ResponseWriter, r *http.Request, prefix string) {
	name := r.PathValue("name")
	kinds := map[string]string{".css": "text/css; charset=utf-8", ".js": "text/javascript; charset=utf-8"}
	if prefix == "vendor/" {
		kinds[".svg"] = "image/svg+xml"
		kinds[".txt"] = "text/plain; charset=utf-8"
		kinds[".json"] = "application/json"
	}
	kind := kinds[filepath.Ext(name)]
	if kind == "" || filepath.Base(name) != name {
		fail(w, 404, "not_found", "asset not found")
		return
	}
	data, err := web.Files.ReadFile(prefix + name)
	if err != nil {
		fail(w, 404, "not_found", "asset not found")
		return
	}
	w.Header().Set("Content-Type", kind)
	_, _ = w.Write(data)
}
