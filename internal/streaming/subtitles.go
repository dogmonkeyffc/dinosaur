package streaming

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"

	"dinosaur/internal/media"
)

var ErrSubtitle = errors.New("subtitle unavailable or unsupported")

type Subtitle struct {
	ID           int    `json:"id"`
	Label        string `json:"label"`
	Kind         string `json:"kind"`
	Delivery     string `json:"delivery"`
	Codec        string `json:"codec,omitempty"`
	Language     string `json:"language,omitempty"`
	RelativePath string `json:"-"`
}

func Subtitles(root, video string, tracks ...media.Track) []Subtitle {
	result := embeddedSubtitles(tracks)
	externalID := 0
	path, err := media.ResolveFile(root, video)
	if err != nil {
		return result
	}
	directory := filepath.Dir(path)
	stem := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
	entries, err := os.ReadDir(directory)
	if err != nil {
		return result
	}
	for _, entry := range entries {
		extension := strings.ToLower(filepath.Ext(entry.Name()))
		if entry.IsDir() || (extension != ".vtt" && extension != ".srt") {
			continue
		}
		name := strings.TrimSuffix(entry.Name(), filepath.Ext(entry.Name()))
		if name != stem && !strings.HasPrefix(name, stem+".") {
			continue
		}
		relative, err := filepath.Rel(root, filepath.Join(directory, entry.Name()))
		if err != nil {
			continue
		}
		if _, err := media.ResolveFile(root, relative); err != nil {
			continue
		}
		label := strings.TrimPrefix(name, stem)
		label = strings.Trim(label, ". ")
		if label == "" {
			label = "外挂字幕"
		}
		result = append(result, Subtitle{ID: externalID, Label: label, Kind: "external", Delivery: "text", RelativePath: relative})
		externalID++
	}
	return result
}
func ReadSubtitle(root, relative string) ([]byte, error) {
	path, err := media.ResolveFile(root, relative)
	if err != nil {
		return nil, ErrSubtitle
	}
	file, err := os.Open(path)
	if err != nil {
		return nil, ErrSubtitle
	}
	defer file.Close()
	data, err := io.ReadAll(io.LimitReader(file, 2*1024*1024+1))
	if err != nil || len(data) > 2*1024*1024 || !utf8.Valid(data) {
		return nil, ErrSubtitle
	}
	content := strings.TrimPrefix(string(data), "\ufeff")
	content = strings.ReplaceAll(content, "\r\n", "\n")
	if strings.EqualFold(filepath.Ext(path), ".srt") {
		content = "WEBVTT\n\n" + srtTimestamp.ReplaceAllString(content, "$1.$2")
	} else if !strings.HasPrefix(content, "WEBVTT") {
		return nil, ErrSubtitle
	}
	return []byte(content), nil
}
