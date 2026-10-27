package streaming

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"sync"
	"time"
)

// SubtitleExtractor bounds both FFmpeg concurrency and the in-memory segment cache.
// Segments overlap so seeking and cues spanning minute boundaries remain readable.
type SubtitleExtractor struct {
	executable string
	slots      chan struct{}
	mu         sync.Mutex
	cache      map[string][]byte
	order      []string
}

func NewSubtitleExtractor(executable string) *SubtitleExtractor {
	if executable == "" {
		executable = "ffmpeg"
	}
	return &SubtitleExtractor{executable: executable, slots: make(chan struct{}, 2), cache: make(map[string][]byte)}
}

func (e *SubtitleExtractor) Read(ctx context.Context, path string, index int, position float64) ([]byte, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	start := max(0, float64(int(position/60)*60)-15)
	key := fmt.Sprintf("%s:%d:%d:%d:%.0f", path, info.Size(), info.ModTime().UnixNano(), index, start)
	e.mu.Lock()
	cached := e.cache[key]
	e.mu.Unlock()
	if cached != nil {
		return cached, nil
	}
	select {
	case e.slots <- struct{}{}:
		defer func() { <-e.slots }()
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		return nil, ErrBusy
	}
	worker, cancel := context.WithTimeout(ctx, 12*time.Second)
	defer cancel()
	args := []string{"-v", "error", "-nostdin", "-threads", "1", "-copyts", "-ss", strconv.FormatFloat(start, 'f', 3, 64), "-protocol_whitelist", "file,pipe", "-i", path, "-to", strconv.FormatFloat(start+105, 'f', 3, 64), "-map", "0:" + strconv.Itoa(index), "-vn", "-an", "-c:s", "webvtt", "-avoid_negative_ts", "disabled", "-f", "webvtt", "pipe:1"}
	command := exec.CommandContext(worker, e.executable, args...)
	command.WaitDelay = time.Second
	output := &subtitleBuffer{}
	command.Stdout = output
	command.Stderr = &stderrTail{}
	if err := command.Run(); err != nil {
		return nil, ErrSubtitle
	}
	if !bytes.HasPrefix(output.Bytes(), []byte("WEBVTT")) {
		return nil, ErrSubtitle
	}
	data, err := normalizeSubtitle(output.Bytes())
	if err != nil {
		return nil, err
	}
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}
	e.mu.Lock()
	if _, exists := e.cache[key]; !exists {
		if len(e.order) >= 24 {
			delete(e.cache, e.order[0])
			e.order = e.order[1:]
		}
		e.order = append(e.order, key)
		e.cache[key] = data
	}
	e.mu.Unlock()
	return data, nil
}

type subtitleBuffer struct{ bytes.Buffer }

func (b *subtitleBuffer) Write(p []byte) (int, error) {
	if b.Len()+len(p) > 2*1024*1024 {
		return 0, ErrSubtitle
	}
	return b.Buffer.Write(p)
}
