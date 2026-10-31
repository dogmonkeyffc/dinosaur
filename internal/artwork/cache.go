// Package artwork stores uploaded artwork without fetching caller-provided URLs.
package artwork

import (
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	_ "image/png"
	"io"
	"os"
	"path/filepath"
)

const MaxSize = 5 * 1024 * 1024

var slots = make(chan struct{}, 2)
var ErrTooLarge = fmt.Errorf("artwork exceeds 5 MiB")
var ErrImage = fmt.Errorf("invalid or oversized JPEG/PNG artwork")

func Cache(ctx context.Context, reader io.Reader, directory string) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	select {
	case slots <- struct{}{}:
		defer func() { <-slots }()
	case <-ctx.Done():
		return "", ctx.Err()
	}
	data, err := io.ReadAll(io.LimitReader(reader, MaxSize+1))
	if len(data) > MaxSize {
		return "", ErrTooLarge
	}
	if err != nil {
		return "", ErrImage
	}
	config, format, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil || (format != "jpeg" && format != "png") || config.Width <= 0 || config.Height <= 0 || config.Width > 4000 || config.Height > 4000 || int64(config.Width)*int64(config.Height) > 8_000_000 {
		return "", ErrImage
	}
	decoded, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return "", ErrImage
	}
	canvas := image.NewRGBA(decoded.Bounds())
	draw.Draw(canvas, canvas.Bounds(), image.NewUniform(color.White), image.Point{}, draw.Src)
	draw.Draw(canvas, canvas.Bounds(), decoded, decoded.Bounds().Min, draw.Over)
	var normalized bytes.Buffer
	if jpeg.Encode(&normalized, canvas, &jpeg.Options{Quality: 90}) != nil {
		return "", ErrImage
	}
	if normalized.Len() > MaxSize {
		return "", ErrImage
	}
	key := fmt.Sprintf("%x", sha256.Sum256(normalized.Bytes()))
	if err := ctx.Err(); err != nil {
		return "", err
	}
	posters := filepath.Join(directory, "posters")
	if err := os.MkdirAll(posters, 0700); err != nil {
		return "", err
	}
	file, err := os.CreateTemp(posters, ".agent-*.jpg")
	if err != nil {
		return "", err
	}
	defer os.Remove(file.Name())
	_, writeErr := file.Write(normalized.Bytes())
	closeErr := file.Close()
	if writeErr != nil {
		return "", writeErr
	}
	if closeErr != nil {
		return "", closeErr
	}
	return key, os.Rename(file.Name(), filepath.Join(posters, key+".jpg"))
}
