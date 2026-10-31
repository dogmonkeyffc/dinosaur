package artwork

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"hash/crc32"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestUploadedArtworkIsValidatedAndNormalized(t *testing.T) {
	var input bytes.Buffer
	if err := png.Encode(&input, image.NewRGBA(image.Rect(0, 0, 12, 20))); err != nil {
		t.Fatal(err)
	}
	directory := t.TempDir()
	key, err := Cache(context.Background(), bytes.NewReader(input.Bytes()), directory)
	if err != nil || len(key) != 64 {
		t.Fatalf("cache: %s %v", key, err)
	}
	path := filepath.Join(directory, "posters", key+".jpg")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	config, format, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil || format != "jpeg" || config.Width != 12 || config.Height != 20 {
		t.Fatal("upload not normalized")
	}
	info, _ := os.Stat(path)
	if info.Mode().Perm() != 0600 {
		t.Fatal("artwork permissions too broad")
	}
	forged := append([]byte(nil), input.Bytes()...)
	binary.BigEndian.PutUint32(forged[16:20], 5000)
	binary.BigEndian.PutUint32(forged[29:33], crc32.ChecksumIEEE(forged[12:29]))
	if _, err := Cache(context.Background(), bytes.NewReader(forged), directory); !errors.Is(err, ErrImage) {
		t.Fatal("oversized dimensions accepted")
	}
	if _, err := Cache(context.Background(), strings.NewReader("invalid image"), directory); !errors.Is(err, ErrImage) {
		t.Fatal("invalid image accepted")
	}
	if _, err := Cache(context.Background(), strings.NewReader(strings.Repeat("x", MaxSize+1)), directory); !errors.Is(err, ErrTooLarge) {
		t.Fatal("oversized body accepted")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := Cache(ctx, bytes.NewReader(input.Bytes()), directory); !errors.Is(err, context.Canceled) {
		t.Fatal("cancelled upload accepted")
	}
}
