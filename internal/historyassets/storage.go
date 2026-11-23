package historyassets

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"os"
	"path/filepath"

	"dinosaur/internal/history"
)

func Put(directory string, data []byte) (string, error) {
	if len(data) == 0 || len(data) > 2*1024*1024 {
		return "", errors.New("invalid history poster size")
	}
	ext := ""
	switch {
	case len(data) > 12 && string(data[:4]) == "RIFF" && string(data[8:12]) == "WEBP":
		ext = ".webp"
	case bytes.HasPrefix(data, []byte{0xff, 0xd8, 0xff}):
		ext = ".jpg"
	case bytes.HasPrefix(data, []byte("\x89PNG\r\n\x1a\n")):
		ext = ".png"
	default:
		return "", errors.New("unsupported history poster")
	}
	sum := sha256.Sum256(data)
	key := hex.EncodeToString(sum[:]) + ext
	if err := os.MkdirAll(directory, 0700); err != nil {
		return "", err
	}
	file, err := os.OpenFile(filepath.Join(directory, key), os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if errors.Is(err, os.ErrExist) {
		root, err := os.OpenRoot(directory)
		if err != nil {
			return "", err
		}
		defer root.Close()
		existing, err := root.ReadFile(key)
		if err != nil {
			return "", err
		}
		if !bytes.Equal(data, existing) {
			return "", errors.New("history poster content mismatch")
		}
		return key, nil
	}
	if err != nil {
		return "", err
	}
	_, err = file.Write(data)
	closeErr := file.Close()
	if err != nil {
		return "", err
	}
	return key, closeErr
}

func Open(directory, key string) (*os.File, error) {
	if !history.ValidPosterKey(key) {
		return nil, os.ErrNotExist
	}
	root, err := os.OpenRoot(directory)
	if err != nil {
		return nil, err
	}
	defer root.Close()
	return root.Open(key)
}
