package streaming

import (
	"context"
	"io"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"dinosaur/internal/playback"
)

func TestDelayedAudioDoesNotBlockVideoStartup(t *testing.T) {
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		t.Skip("FFmpeg is not installed")
	}
	path := filepath.Join(t.TempDir(), "delayed.mkv")
	command := exec.Command("ffmpeg", "-v", "error", "-f", "lavfi", "-i", "color=c=green:s=64x48:d=60", "-itsoffset", "52", "-f", "lavfi", "-i", "sine=frequency=440:duration=8", "-c:v", "libx264", "-preset", "ultrafast", "-threads", "1", "-c:a", "aac", path)
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("fixture: %s %v", output, err)
	}
	transcoder := New("ffmpeg")
	defer transcoder.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	stream, err := transcoder.Open(ctx, path, playback.Plan{Mode: "transcode", Audio: 1, Start: 2})
	if err != nil {
		t.Fatal(err)
	}
	defer stream.Close()
	if _, err := io.ReadFull(stream, make([]byte, 32)); err != nil {
		stream.Close()
		t.Fatalf("delayed audio blocked video startup: %v %s", err, stream.Diagnostics())
	}
	stream.Ready()
}
