package server

import (
	"encoding/json"
	"net/url"
	"os"
	"path/filepath"
	"testing"
)

func TestLibraryDirectoryPickerListsFoldersOnly(t *testing.T) {
	s, _ := fixture(t)
	setupAdmin(t, s)
	admin := loginAs(t, s, "admin", testPassword)
	root := t.TempDir()
	folder := filepath.Join(root, "动漫")
	if err := os.Mkdir(folder, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "sample.mkv"), nil, 0600); err != nil {
		t.Fatal(err)
	}
	alias := filepath.Join(root, "alias")
	if err := os.Symlink(folder, alias); err != nil {
		t.Fatal(err)
	}
	const endpoint = "/api/admin/libraries/directories?path="
	status(t, request(t, s, "GET", endpoint+url.QueryEscape(root), nil, nil, nil), 401)
	status(t, request(t, s, "GET", endpoint+"relative", nil, &admin, nil), 400)
	status(t, request(t, s, "GET", endpoint+url.QueryEscape(alias), nil, &admin, nil), 404)
	response := request(t, s, "GET", endpoint+url.QueryEscape(root), nil, &admin, nil)
	status(t, response, 200)
	var result struct {
		Path, Parent string
		Folders      []struct{ Name, Path string }
	}
	if err := json.Unmarshal(response.Body.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	if result.Path != root || result.Parent != filepath.Dir(root) || len(result.Folders) != 1 ||
		result.Folders[0].Path != folder || result.Folders[0].Name != "动漫" {
		t.Fatalf("unexpected listing: %+v", result)
	}
}
