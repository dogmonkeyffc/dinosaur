package config

import (
	"fmt"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Addr, DataDir, DatabasePath, FFmpegPath, FFprobePath string
	SetupToken, PublicURL                                string
	CookieSecure                                         bool
	SessionTTL, RequestTimeout, ShutdownTimeout          time.Duration
}

func Load() (Config, error) {
	dataDir, configured := os.LookupEnv("DINOSAUR_DATA_DIR")
	if !configured {
		home, err := os.UserHomeDir()
		if err != nil {
			return Config{}, fmt.Errorf("resolve home directory: %w", err)
		}
		dataDir = filepath.Join(home, ".cache", "dinosaur")
	}
	c := Config{
		Addr:        env("DINOSAUR_ADDR", "127.0.0.1:8096"),
		DataDir:     dataDir,
		FFmpegPath:  env("DINOSAUR_FFMPEG_PATH", "ffmpeg"),
		FFprobePath: env("DINOSAUR_FFPROBE_PATH", "ffprobe"),
		SetupToken:  os.Getenv("DINOSAUR_SETUP_TOKEN"),
		PublicURL:   strings.TrimRight(os.Getenv("DINOSAUR_PUBLIC_URL"), "/"),
	}
	c.DatabasePath = env("DINOSAUR_DATABASE_PATH", filepath.Join(c.DataDir, "dinosaur.db"))
	var err error
	c.CookieSecure, err = strconv.ParseBool(env("DINOSAUR_COOKIE_SECURE", "false"))
	if err != nil {
		return c, fmt.Errorf("DINOSAUR_COOKIE_SECURE must be a boolean")
	}
	for _, field := range []struct {
		name, fallback string
		target         *time.Duration
	}{
		{"DINOSAUR_SESSION_TTL", "168h", &c.SessionTTL},
		{"DINOSAUR_REQUEST_TIMEOUT", "15s", &c.RequestTimeout},
		{"DINOSAUR_SHUTDOWN_TIMEOUT", "10s", &c.ShutdownTimeout},
	} {
		*field.target, err = time.ParseDuration(env(field.name, field.fallback))
		if err != nil || *field.target < time.Second || *field.target > 365*24*time.Hour {
			return c, fmt.Errorf("%s must be between 1s and 8760h", field.name)
		}
	}
	if _, _, err := net.SplitHostPort(c.Addr); err != nil {
		return c, fmt.Errorf("DINOSAUR_ADDR: %w", err)
	}
	_, port, _ := net.SplitHostPort(c.Addr)
	if n, err := strconv.Atoi(port); err != nil || n < 1 || n > 65535 {
		return c, fmt.Errorf("DINOSAUR_ADDR port must be between 1 and 65535")
	}
	for name, value := range map[string]string{"DINOSAUR_DATA_DIR": c.DataDir, "DINOSAUR_DATABASE_PATH": c.DatabasePath, "DINOSAUR_FFMPEG_PATH": c.FFmpegPath, "DINOSAUR_FFPROBE_PATH": c.FFprobePath} {
		if strings.TrimSpace(value) == "" {
			return c, fmt.Errorf("%s must not be empty", name)
		}
	}
	if c.DatabasePath == ":memory:" {
		return c, fmt.Errorf("DINOSAUR_DATABASE_PATH must be a persistent file path")
	}
	if c.SetupToken != "" && len(c.SetupToken) < 32 {
		return c, fmt.Errorf("DINOSAUR_SETUP_TOKEN must contain at least 32 bytes")
	}
	if c.PublicURL != "" {
		u, err := url.Parse(c.PublicURL)
		if err != nil || u.Host == "" || (u.Scheme != "http" && u.Scheme != "https") || u.User != nil || u.Path != "" || u.RawQuery != "" || u.Fragment != "" {
			return c, fmt.Errorf("DINOSAUR_PUBLIC_URL must be an http(s) origin without a path")
		}
		if (u.Scheme == "https") != c.CookieSecure {
			return c, fmt.Errorf("DINOSAUR_COOKIE_SECURE must match DINOSAUR_PUBLIC_URL scheme")
		}
	}
	return c, nil
}

func env(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
