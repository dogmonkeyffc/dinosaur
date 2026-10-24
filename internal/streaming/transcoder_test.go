package streaming

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"dinosaur/internal/playback"
)

func TestFFmpegStreamCancellationAndCapacity(t *testing.T) {
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		t.Skip("FFmpeg is not installed")
	}
	path := filepath.Join(t.TempDir(), "video.mkv")
	command := exec.Command("ffmpeg", "-v", "error", "-f", "lavfi", "-i", "color=c=green:s=64x48:d=30", "-c:v", "libx264", "-preset", "ultrafast", "-threads", "1", path)
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("fixture: %s %v", output, err)
	}
	transcoder := New("ffmpeg")
	defer transcoder.Close()
	plan := playback.Plan{Mode: "remux", Audio: -1}
	first, err := transcoder.Open(context.Background(), path, plan)
	if err != nil {
		t.Fatal(err)
	}
	defer first.Close()
	header := make([]byte, 32)
	if _, err := io.ReadFull(first, header); err != nil {
		t.Fatal(err)
	}
	first.Ready()
	if !bytes.Contains(header, []byte("ftyp")) {
		t.Fatal("not an MP4 stream")
	}
	second, err := transcoder.Open(context.Background(), path, plan)
	if err != nil {
		t.Fatal(err)
	}
	defer second.Close()
	limited, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	if _, err := transcoder.Open(limited, path, plan); !errors.Is(err, ErrBusy) {
		t.Fatalf("capacity limit=%v", err)
	}
	if err := first.Close(); err != nil {
		t.Fatal(err)
	}
	if err := second.Close(); err != nil {
		t.Fatal(err)
	}
	if len(transcoder.slots) != 0 {
		t.Fatal("stream slots not released")
	}
	transcoder.Close()
	if _, err := transcoder.Open(context.Background(), path, plan); err == nil {
		t.Fatal("stream started after shutdown")
	}
}
