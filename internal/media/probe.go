package media

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"math"
	"os/exec"
	"strconv"
	"time"
)

type ProbeResult struct {
	Duration      float64
	Format        string
	VideoCodec    string
	Width, Height int
	Tracks        []Track
}

type boundedBuffer struct {
	data  []byte
	limit int
}

func (b *boundedBuffer) Write(p []byte) (int, error) {
	if len(b.data)+len(p) > b.limit {
		return 0, errors.New("media tool output exceeds limit")
	}
	b.data = append(b.data, p...)
	return len(p), nil
}

func Probe(ctx context.Context, executable, path string) (ProbeResult, error) {
	if executable == "" {
		executable = "ffprobe"
	}
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, executable, "-v", "error", "-protocol_whitelist", "file,pipe", "-show_entries", "format=duration,format_name:stream=index,codec_type,codec_name,width,height,channels:stream_tags=language,title:stream_disposition=attached_pic", "-of", "json", path)
	cmd.WaitDelay = time.Second
	output := &boundedBuffer{limit: 1024 * 1024}
	cmd.Stdout = output
	cmd.Stderr = io.Discard
	if err := cmd.Run(); err != nil {
		return ProbeResult{}, err
	}
	var data struct {
		Format struct {
			Duration string `json:"duration"`
			Name     string `json:"format_name"`
		} `json:"format"`
		Streams []struct {
			Index    int    `json:"index"`
			Type     string `json:"codec_type"`
			Codec    string `json:"codec_name"`
			Width    int    `json:"width"`
			Height   int    `json:"height"`
			Channels int    `json:"channels"`
			Tags     struct {
				Language string `json:"language"`
				Title    string `json:"title"`
			} `json:"tags"`
			Disposition struct {
				AttachedPicture int `json:"attached_pic"`
			} `json:"disposition"`
		} `json:"streams"`
	}
	if err := json.Unmarshal(output.data, &data); err != nil {
		return ProbeResult{}, err
	}
	result := ProbeResult{Format: data.Format.Name, Tracks: make([]Track, 0)}
	result.Duration, _ = strconv.ParseFloat(data.Format.Duration, 64)
	if math.IsNaN(result.Duration) || math.IsInf(result.Duration, 0) || result.Duration < 0 {
		result.Duration = 0
	}
	for _, stream := range data.Streams {
		if stream.Type == "video" && result.VideoCodec == "" && stream.Disposition.AttachedPicture == 0 {
			result.VideoCodec = stream.Codec
			result.Width = stream.Width
			result.Height = stream.Height
		}
		if stream.Type == "audio" || stream.Type == "subtitle" {
			result.Tracks = append(result.Tracks, Track{Index: stream.Index, Type: stream.Type, Codec: stream.Codec, Channels: stream.Channels, Language: stream.Tags.Language, Title: stream.Tags.Title})
		}
	}
	if result.VideoCodec == "" {
		return result, errors.New("no video stream")
	}
	return result, nil
}
