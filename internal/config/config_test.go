package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func clearEnv(t *testing.T) {
	t.Helper()
	defaults := map[string]string{"DINOSAUR_ADDR": "127.0.0.1:8096", "DINOSAUR_DATA_DIR": "./data", "DINOSAUR_DATABASE_PATH": "./data/dinosaur.db", "DINOSAUR_FFMPEG_PATH": "ffmpeg", "DINOSAUR_FFPROBE_PATH": "ffprobe", "DINOSAUR_SETUP_TOKEN": "", "DINOSAUR_PUBLIC_URL": "", "DINOSAUR_COOKIE_SECURE": "false", "DINOSAUR_SESSION_TTL": "168h", "DINOSAUR_REQUEST_TIMEOUT": "15s", "DINOSAUR_SHUTDOWN_TIMEOUT": "10s"}
	for k, v := range defaults {
		t.Setenv(k, v)
	}
}
func TestLoad(t *testing.T) {
	clearEnv(t)
	c, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if c.SessionTTL != 7*24*time.Hour || c.CookieSecure {
		t.Fatalf("bad defaults %+v", c)
	}
	for _, tc := range []struct{ key, value string }{
		{"DINOSAUR_ADDR", "invalid"}, {"DINOSAUR_ADDR", "localhost:99999"}, {"DINOSAUR_SESSION_TTL", "0s"}, {"DINOSAUR_SESSION_TTL", "not-a-duration"}, {"DINOSAUR_REQUEST_TIMEOUT", "-1s"}, {"DINOSAUR_COOKIE_SECURE", "maybe"}, {"DINOSAUR_DATA_DIR", ""}, {"DINOSAUR_DATABASE_PATH", ":memory:"}, {"DINOSAUR_SETUP_TOKEN", "short"}, {"DINOSAUR_PUBLIC_URL", "https://example.com/path"}, {"DINOSAUR_PUBLIC_URL", "https://example.com"},
	} {
		t.Run(tc.key+"="+tc.value, func(t *testing.T) {
			t.Setenv(tc.key, tc.value)
			if _, err := Load(); err == nil {
				t.Fatal("invalid config accepted")
			}
		})
	}
	t.Setenv("DINOSAUR_COOKIE_SECURE", "true")
	t.Setenv("DINOSAUR_PUBLIC_URL", "https://media.example.com")
	if _, err := Load(); err != nil {
		t.Fatal(err)
	}
}

func TestDefaultDataDirectory(t *testing.T) {
	clearEnv(t)
	if err := os.Unsetenv("DINOSAUR_DATA_DIR"); err != nil {
		t.Fatal(err)
	}
	if err := os.Unsetenv("DINOSAUR_DATABASE_PATH"); err != nil {
		t.Fatal(err)
	}
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatal(err)
	}
	c, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Join(home, ".cache", "dinosaur")
	if c.DataDir != want || c.DatabasePath != filepath.Join(want, "dinosaur.db") {
		t.Fatalf("unexpected defaults: data=%q database=%q", c.DataDir, c.DatabasePath)
	}
	custom := t.TempDir()
	t.Setenv("DINOSAUR_DATA_DIR", custom)
	c, err = Load()
	if err != nil {
		t.Fatal(err)
	}
	if c.DataDir != custom || c.DatabasePath != filepath.Join(custom, "dinosaur.db") {
		t.Fatal("database did not follow the custom data directory")
	}
}
