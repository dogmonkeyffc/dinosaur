package media

import (
	"context"
	"crypto/sha256"
	"fmt"
	"image"
	"image/jpeg"
	_ "image/png"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"time"
)

func PosterKey(item Item) string {
	sum := sha256.Sum256([]byte(fmt.Sprintf("%d:%s:%d:%d", item.LibraryID, item.RelativePath, item.Size, item.ModifiedNS)))
	return fmt.Sprintf("%x", sum)
}

// BuildPoster writes only to the application cache. A sibling image is preferred
// to a frame extracted near the start of the video.
func BuildPoster(ctx context.Context, executable, cache, root, path string, item Item) (string, error) {
	key := PosterKey(item)
	directory := filepath.Join(cache, "posters")
	if err := os.MkdirAll(directory, 0700); err != nil {
		return "", err
	}
	destination := filepath.Join(directory, key+".jpg")
	temporary, err := os.CreateTemp(directory, ".poster-*.jpg")
	if err != nil {
		return "", err
	}
	tempPath := temporary.Name()
	if err := temporary.Close(); err != nil {
		return "", err
	}
	defer os.Remove(tempPath)
	if err := copyLocalPoster(root, path, tempPath); err != nil {
		if executable == "" {
			executable = "ffmpeg"
		}
		ctx, cancel := context.WithTimeout(ctx, 20*time.Second)
		defer cancel()
		position := 30.0
		if item.Duration > 0 {
			position = min(item.Duration/10, 300)
		}
		cmd := exec.CommandContext(ctx, executable, "-hide_banner", "-loglevel", "error", "-nostdin", "-y", "-threads", "1", "-ss", strconv.FormatFloat(position, 'f', 2, 64), "-protocol_whitelist", "file,pipe", "-i", path, "-map", "0:V:0", "-frames:v", "1", "-vf", "scale=640:-2", "-threads", "1", "-q:v", "3", tempPath)
		cmd.WaitDelay = time.Second
		cmd.Stdout = io.Discard
		cmd.Stderr = io.Discard
		if err := cmd.Run(); err != nil {
			return "", err
		}
	}
	info, err := os.Stat(tempPath)
	if err != nil {
		return "", err
	}
	if info.Size() == 0 {
		return "", fmt.Errorf("empty poster")
	}
	if err := os.Rename(tempPath, destination); err != nil {
		return "", err
	}
	return key, nil
}

func copyLocalPoster(root, video, destination string) error {
	stem := video[:len(video)-len(filepath.Ext(video))]
	candidates := []string{stem + "-poster.jpg", stem + ".jpg", stem + ".png"}
	if filepath.Dir(video) != root {
		for _, name := range []string{"poster.jpg", "poster.png", "folder.jpg", "cover.jpg"} {
			candidates = append(candidates, filepath.Join(filepath.Dir(video), name))
		}
	}
	for _, candidate := range candidates {
		relative, err := filepath.Rel(root, candidate)
		if err != nil {
			continue
		}
		path, err := ResolveFile(root, relative)
		if err != nil {
			continue
		}
		if err := encodeLocalImage(path, destination); err == nil {
			return nil
		}
	}
	return fmt.Errorf("no usable local poster")
}
func encodeLocalImage(path, destination string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil {
		return err
	}
	if info.Size() > 20*1024*1024 {
		return fmt.Errorf("poster too large")
	}
	config, _, err := image.DecodeConfig(file)
	if err != nil {
		return err
	}
	if config.Width <= 0 || config.Height <= 0 || int64(config.Width)*int64(config.Height) > 16_000_000 {
		return fmt.Errorf("poster dimensions too large")
	}
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return err
	}
	decoded, _, err := image.Decode(file)
	if err != nil {
		return err
	}
	output, err := os.OpenFile(destination, os.O_WRONLY|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	encodeErr := jpeg.Encode(output, decoded, &jpeg.Options{Quality: 85})
	closeErr := output.Close()
	if encodeErr != nil {
		return encodeErr
	}
	return closeErr
}
