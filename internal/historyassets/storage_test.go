package historyassets

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPostersAreContentAddressedAndCannotEscapeTheirDirectory(t *testing.T) {
	dir := t.TempDir()
	data := []byte("RIFF0000WEBPposter-data")
	key, err := Put(dir, data)
	if err != nil {
		t.Fatal(err)
	}
	second, err := Put(dir, data)
	if err != nil || key != second {
		t.Fatal("poster write is not idempotent")
	}
	file, err := Open(dir, key)
	if err != nil {
		t.Fatal(err)
	}
	file.Close()
	if _, err := Put(dir, []byte("<html>not an image</html>")); err == nil {
		t.Fatal("non-image accepted")
	}
	if _, err := Open(dir, "../private"); err == nil {
		t.Fatal("path traversal accepted")
	}
	outside := filepath.Join(t.TempDir(), "secret")
	if err := os.WriteFile(outside, []byte("private"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(filepath.Join(dir, key)); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, filepath.Join(dir, key)); err != nil {
		t.Fatal(err)
	}
	if _, err := Open(dir, key); err == nil {
		t.Fatal("poster symlink escaped root")
	}
}
