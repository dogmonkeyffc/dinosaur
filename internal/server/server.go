package server

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"mime"
	"net"
	"net/http"
	"strconv"
	"time"

	"dinosaur/internal/auth"
	"dinosaur/internal/config"
	"dinosaur/internal/scanner"
	"dinosaur/internal/scraping"
	"dinosaur/internal/store"
	"dinosaur/internal/streaming"
	"dinosaur/internal/tmdb"
)

const cookieName = "dinosaur_session"

type identityKey struct{}
type identity struct {
	user  store.User
	token string
}
type Server struct {
	config     config.Config
	store      *store.Store
	logger     *slog.Logger
	hasher     *auth.Hasher
	limiter    *auth.Limiter
	dummyHash  string
	handler    http.Handler
	scanner    *scanner.Manager
	transcoder *streaming.Transcoder
	tmdb       *tmdb.Client
	scraper    *scraping.Manager
}

func New(c config.Config, db *store.Store, logger *slog.Logger) (*Server, error) {
	h := auth.NewHasher()
	random, err := auth.NewToken()
	if err != nil {
		return nil, err
	}
	dummy, err := h.Hash(context.Background(), random)
	if err != nil {
		return nil, err
	}
	s := &Server{config: c, store: db, logger: logger, hasher: h, limiter: auth.NewLimiter(), dummyHash: dummy}
	s.scanner = scanner.New(db, c, logger)
	s.transcoder = streaming.New(c.FFmpegPath)
	s.tmdb = tmdb.New()
	s.scraper = scraping.New(db, s.tmdb, c.DataDir)
	s.scanner.AfterScan = s.scraper.Schedule
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health/live", func(w http.ResponseWriter, r *http.Request) { writeJSON(w, 200, map[string]string{"status": "ok"}) })
	mux.HandleFunc("GET /health/ready", s.ready)
	mux.HandleFunc("GET /api/setup", s.setupStatus)
	mux.HandleFunc("POST /api/setup", s.setup)
	mux.HandleFunc("POST /api/auth/login", s.login)
	mux.Handle("GET /api/auth/me", s.require(false, http.HandlerFunc(s.me)))
	mux.Handle("POST /api/auth/logout", s.require(false, http.HandlerFunc(s.logout)))
	mux.Handle("POST /api/auth/password", s.require(false, http.HandlerFunc(s.changePassword)))
	mux.Handle("POST /api/admin/users", s.require(true, http.HandlerFunc(s.createUser)))
	mux.Handle("PATCH /api/admin/users/{id}", s.require(true, http.HandlerFunc(s.updateUser)))
	mux.Handle("GET /api/libraries", s.require(false, http.HandlerFunc(s.listLibraries)))
	mux.Handle("GET /api/admin/libraries", s.require(true, http.HandlerFunc(s.listLibraries)))
	mux.Handle("POST /api/admin/libraries", s.require(true, http.HandlerFunc(s.saveLibrary)))
	mux.Handle("PUT /api/admin/libraries/{id}", s.require(true, http.HandlerFunc(s.saveLibrary)))
	mux.Handle("DELETE /api/admin/libraries/{id}", s.require(true, http.HandlerFunc(s.deleteLibrary)))
	mux.Handle("POST /api/admin/libraries/{id}/scan", s.require(true, http.HandlerFunc(s.startScan)))
	mux.Handle("GET /api/admin/libraries/{id}/scan", s.require(true, http.HandlerFunc(s.scanStatus)))
	mux.Handle("GET /api/media", s.require(false, http.HandlerFunc(s.listMedia)))
	mux.Handle("GET /api/media/{id}", s.require(false, http.HandlerFunc(s.getMedia)))
	mux.Handle("GET /api/media/{id}/poster", s.require(false, http.HandlerFunc(s.mediaPoster)))
	s.mountPlayback(mux)
	s.mountMetadataSettings(mux)
	s.mountScraping(mux)
	if err := s.mountWeb(mux); err != nil {
		return nil, err
	}
	// These boundaries also cover future media, artwork, subtitles and playback routes.
	for _, prefix := range []string{"/api/media/", "/api/playback/"} {
		mux.Handle(prefix, s.require(false, http.HandlerFunc(s.notImplemented)))
	}
	mux.Handle("/api/admin/", s.require(true, http.HandlerFunc(s.notImplemented)))
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) { fail(w, 404, "not_found", "route not found") })
	timeoutBody := `{"error":{"code":"request_timeout","message":"request timed out"}}` + "\n"
	routes := http.NewServeMux()
	routes.Handle("GET /api/playback/{id}/stream", s.require(false, http.HandlerFunc(s.servePlaybackStream)))
	routes.Handle("/", http.TimeoutHandler(mux, c.RequestTimeout, timeoutBody))
	s.handler = s.observe(s.checkOrigin(routes))
	return s, nil
}
func (s *Server) Handler() http.Handler { return s.handler }
func (s *Server) ready(w http.ResponseWriter, r *http.Request) {
	if err := s.store.DB.PingContext(r.Context()); err != nil {
		fail(w, 503, "not_ready", "database unavailable")
		return
	}
	writeJSON(w, 200, map[string]string{"status": "ready"})
}
func (s *Server) notImplemented(w http.ResponseWriter, r *http.Request) {
	fail(w, 501, "not_implemented", "media functionality is planned for a later phase")
}
func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
func fail(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, map[string]any{"error": map[string]string{"code": code, "message": message}})
}
func (s *Server) internal(w http.ResponseWriter, r *http.Request, err error) {
	s.logger.ErrorContext(r.Context(), "request failed", "request_id", w.Header().Get("X-Request-ID"), "error", err)
	fail(w, 500, "internal_error", "internal server error")
}
func decodeJSON(w http.ResponseWriter, r *http.Request, dst any) bool {
	kind, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil || kind != "application/json" {
		fail(w, 415, "unsupported_media_type", "Content-Type must be application/json")
		return false
	}
	r.Body = http.MaxBytesReader(w, r.Body, 16*1024)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	var raw json.RawMessage
	if err = decoder.Decode(&raw); err != nil {
		bodyError(w, err)
		return false
	}
	if len(raw) == 0 || raw[0] != '{' {
		bodyError(w, nil)
		return false
	}
	strict := json.NewDecoder(bytes.NewReader(raw))
	strict.DisallowUnknownFields()
	if err := strict.Decode(dst); err != nil {
		bodyError(w, err)
		return false
	}
	var extra any
	if err = decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		bodyError(w, err)
		return false
	}
	return true
}
func bodyError(w http.ResponseWriter, err error) {
	var tooLarge *http.MaxBytesError
	if errors.As(err, &tooLarge) {
		fail(w, 413, "body_too_large", "request body exceeds 16 KiB")
		return
	}
	fail(w, 400, "invalid_json", "body must contain one valid JSON object with supported fields")
}
func peerIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
func (s *Server) rate(w http.ResponseWriter, limits ...auth.Limit) bool {
	if s.limiter.Allow(limits...) {
		return true
	}
	w.Header().Set("Retry-After", strconv.Itoa(15*60))
	fail(w, 429, "rate_limited", "too many attempts; try again later")
	return false
}
func (s *Server) setCookie(w http.ResponseWriter, token string, expires time.Time) {
	http.SetCookie(w, &http.Cookie{Name: cookieName, Value: token, Path: "/", HttpOnly: true, Secure: s.config.CookieSecure, SameSite: http.SameSiteLaxMode, Expires: expires, MaxAge: int(s.config.SessionTTL.Seconds())})
}
func (s *Server) clearCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{Name: cookieName, Value: "", Path: "/", HttpOnly: true, Secure: s.config.CookieSecure, SameSite: http.SameSiteLaxMode, Expires: time.Unix(1, 0), MaxAge: -1})
}
