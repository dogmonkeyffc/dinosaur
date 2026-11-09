package server

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"dinosaur/internal/media"
)

func TestAgentListsUnindexedLibraryDirectories(t *testing.T) {
	s, db, admin, movies, tv := agentFixture(t)
	_, key := createTestAgentKey(t, s, &admin, tv, "read")
	lib, err := db.LibraryByID(context.Background(), tv)
	if err != nil {
		t.Fatal(err)
	}
	folder := filepath.Join(lib.RootPath, "未扫描")
	if err := os.Mkdir(folder, 0700); err != nil {
		t.Fatal(err)
	}
	for i := range 101 {
		if err := os.WriteFile(filepath.Join(folder, fmt.Sprintf("S01E%03d.mkv", i+1)), nil, 0600); err != nil {
			t.Fatal(err)
		}
	}
	for _, name := range []string{"subtitle.srt", ".hidden.mkv"} {
		if err := os.WriteFile(filepath.Join(folder, name), nil, 0600); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.Symlink(t.TempDir(), filepath.Join(lib.RootPath, "outside")); err != nil {
		t.Fatal(err)
	}
	libraries := agentRequest(t, s, "GET", "/api/agent/libraries", nil, key, "")
	status(t, libraries, 200)
	if !strings.Contains(libraries.Body.String(), `"name":"TV"`) || strings.Contains(libraries.Body.String(), `"name":"Movies"`) || strings.Contains(libraries.Body.String(), lib.RootPath) {
		t.Fatalf("unexpected libraries: %s", libraries.Body.String())
	}
	base := "/api/agent/libraries/" + stringID(tv) + "/files"
	status(t, agentRequest(t, s, "GET", base, nil, "", ""), 401)
	status(t, agentRequest(t, s, "GET", "/api/agent/libraries/"+stringID(movies)+"/files", nil, key, ""), 403)
	status(t, agentRequest(t, s, "GET", base+"?directory=..%2F", nil, key, ""), 400)
	status(t, agentRequest(t, s, "GET", base+"?directory=outside", nil, key, ""), 404)
	read := func(query string) ([]media.DirectoryEntry, string) {
		response := agentRequest(t, s, "GET", base+query, nil, key, "")
		status(t, response, 200)
		var page struct {
			Items []media.DirectoryEntry `json:"items"`
			Next  string                 `json:"next_after"`
		}
		if err := json.Unmarshal(response.Body.Bytes(), &page); err != nil {
			t.Fatal(err)
		}
		return page.Items, page.Next
	}
	items, next := read("")
	if len(items) != 1 || items[0].Path != "未扫描" || items[0].Kind != "directory" || next != "" {
		t.Fatalf("root = %v, cursor = %q", items, next)
	}
	items, next = read("?directory=未扫描")
	if len(items) != 100 || items[0].Path != "未扫描/S01E001.mkv" || items[0].Kind != "video" || next != "S01E100.mkv" {
		t.Fatalf("count = %d, cursor = %q", len(items), next)
	}
	items, next = read("?directory=未扫描&after=" + next)
	if len(items) != 1 || items[0].Path != "未扫描/S01E101.mkv" || next != "" {
		t.Fatalf("last page = %v, cursor = %q", items, next)
	}
}
