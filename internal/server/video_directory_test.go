package server

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"dinosaur/internal/library"
	"dinosaur/internal/media"
	"dinosaur/internal/store"
)

func TestVideoDirectoryShowsActualChildrenAndRespectsLibraryAccess(t *testing.T) {
	s, db := fixture(t)
	setupAdmin(t, s)
	admin := loginAs(t, s, "admin", testPassword)
	user, err := db.UserByName(context.Background(), "admin")
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	if _, err := db.CreateUser(ctx, "viewer", user.PasswordHash, "user", false); err != nil {
		t.Fatal(err)
	}
	viewer := loginAs(t, s, "viewer", testPassword)
	root := t.TempDir()
	for _, dir := range []string{"Take_1%", "empty"} {
		if err := os.Mkdir(filepath.Join(root, dir), 0700); err != nil {
			t.Fatal(err)
		}
	}
	lib, err := db.SaveLibrary(ctx, store.Library{Name: "Videos", Type: "video", RootPath: root})
	if err != nil {
		t.Fatal(err)
	}
	posterKey := strings.Repeat("a", 64)
	posters := filepath.Join(s.config.DataDir, "posters")
	if err := os.MkdirAll(posters, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(posters, posterKey+".jpg"), []byte("test preview"), 0600); err != nil {
		t.Fatal(err)
	}
	for _, relative := range []string{"Root.mp4", filepath.Join("Take_1%", "Raw clip.mp4"), "Missing.mp4"} {
		if relative != "Missing.mp4" {
			if err := os.WriteFile(filepath.Join(root, relative), nil, 0600); err != nil {
				t.Fatal(err)
			}
		}
		item := media.Item{LibraryID: lib.ID, RelativePath: relative, Title: "Indexed title"}
		if strings.Contains(relative, "Raw clip") {
			item.PosterKey = posterKey
		}
		if err := db.SaveMedia(ctx, item, 1); err != nil {
			t.Fatal(err)
		}
	}
	base := fmt.Sprintf("/api/libraries/%d/videos", lib.ID)
	read := func(query string) library.VideoDirectory {
		t.Helper()
		w := request(t, s, "GET", base+query, nil, &admin, nil)
		status(t, w, 200)
		var folder library.VideoDirectory
		if err := json.Unmarshal(w.Body.Bytes(), &folder); err != nil {
			t.Fatal(err)
		}
		return folder
	}
	if value := read(""); len(value.Folders) != 2 || len(value.Videos) != 1 || value.Videos[0].Name != "Root.mp4" || value.Videos[0].PosterURL != "" {
		t.Fatalf("unexpected root: %+v", value)
	}
	nested := read("?path=Take_1%25")
	if nested.Path != "Take_1%" || len(nested.Videos) != 1 || nested.Videos[0].Name != "Raw clip.mp4" {
		t.Fatalf("unexpected child: %+v", nested)
	}
	preview := nested.Videos[0].PosterURL
	expected := fmt.Sprintf("/api/media/%d/poster?v=%s", nested.Videos[0].ID, posterKey)
	if preview != expected {
		t.Fatalf("missing preview: %q", preview)
	}
	image := request(t, s, "GET", preview, nil, &admin, nil)
	if image.Code != 200 || image.Header().Get("Content-Type") != "image/jpeg" || image.Body.String() != "test preview" {
		t.Fatalf("unexpected preview response: %d", image.Code)
	}
	if located := read(fmt.Sprintf("?media_id=%d", nested.Videos[0].ID)); located.Path != nested.Path {
		t.Fatalf("wrong directory for media: %+v", located)
	}
	status(t, request(t, s, "GET", base+"?path=../", nil, &admin, nil), 400)
	if err := db.SaveLibraryAccess(ctx, lib.ID, store.LibraryAccess{Restricted: true}); err != nil {
		t.Fatal(err)
	}
	status(t, request(t, s, "GET", base, nil, &viewer, nil), 404)
	status(t, request(t, s, "GET", preview, nil, &viewer, nil), 404)
	status(t, request(t, s, "GET", base, nil, nil, nil), 401)
}
