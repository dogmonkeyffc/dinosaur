package streaming

import (
	"math"
	"strconv"
	"strings"
	"sync"
)

// ffmpegProgress separates FFmpeg's machine-readable progress from error output.
// Retain only the latest useful speed and bound memory for incomplete lines.
type ffmpegProgress struct {
	mu          sync.Mutex
	line        []byte
	discard     bool
	speed       float64
	diagnostics *stderrTail
}

func (p *ffmpegProgress) Write(data []byte) (int, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	for _, b := range data {
		if b == '\n' {
			if !p.discard {
				p.accept(string(p.line))
			}
			p.line = p.line[:0]
			p.discard = false
		} else if len(p.line) >= 8192 {
			p.discard = true
		} else if !p.discard {
			p.line = append(p.line, b)
		}
	}
	return len(data), nil
}

func (p *ffmpegProgress) accept(line string) {
	key, value, _ := strings.Cut(strings.TrimSpace(line), "=")
	switch key {
	case "speed":
		if raw, ok := strings.CutSuffix(strings.TrimSpace(value), "x"); ok {
			if speed, err := strconv.ParseFloat(raw, 64); err == nil && speed > 0 && !math.IsInf(speed, 0) {
				p.speed = speed
			}
		}
	case "frame", "fps", "bitrate", "total_size", "out_time_us", "out_time_ms", "out_time", "dup_frames", "drop_frames", "progress":
	default:
		if !strings.HasPrefix(key, "stream_") {
			_, _ = p.diagnostics.Write([]byte(line + "\n"))
		}
	}
}

func (stream *Stream) TranscodeSpeed() float64 {
	p := stream.progress
	if p == nil {
		return 0
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.speed
}
