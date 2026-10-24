package media

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseFilename(t *testing.T) {
	cases := []struct {
		path, title string
		year        int
	}{
		{"The.Chorus.2004.BluRay.1080p.x265.mkv", "The Chorus", 2004},
		{"Canola.2016.1080p.friDay.WEB-DL.mkv", "Canola", 2016},
		{"电影.2024.2160p.mkv", "电影", 2024},
		{"A.Movie.1080p.WEB-DL.mp4", "A Movie", 0},
		{"未知片名.mkv", "未知片名", 0},
	}
	for _, test := range cases {
		title, year := ParseFilename(test.path)
		if title != test.title || year != test.year {
			t.Errorf("%s => %s %d", test.path, title, year)
		}
	}
}
func TestResolveFile(t *testing.T) {
	root := t.TempDir()
	safe := filepath.Join(root, "movie.mkv")
	if err := os.WriteFile(safe, []byte("video"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := ResolveFile(root, "movie.mkv"); err != nil {
		t.Fatal(err)
	}
	external := filepath.Join(t.TempDir(), "external.jpg")
	if err := os.WriteFile(external, []byte("secret"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(external, filepath.Join(root, "poster.jpg")); err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{"../outside", safe, "poster.jpg", "."} {
		if _, err := ResolveFile(root, path); err == nil {
			t.Errorf("unsafe path accepted: %s", path)
		}
	}
}
