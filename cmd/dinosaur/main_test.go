package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSetupTokenPersistsPrivately(t *testing.T) {
	path := filepath.Join(t.TempDir(), "setup-token")
	a, err := setupToken(path)
	if err != nil {
		t.Fatal(err)
	}
	b, err := setupToken(path)
	if err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if a != b || len(a) != 43 || info.Mode().Perm() != 0600 {
		t.Fatal("invalid token persistence")
	}
	if err := os.Chmod(path, 0644); err != nil {
		t.Fatal(err)
	}
	if _, err := setupToken(path); err == nil {
		t.Fatal("public token file accepted")
	}
}
