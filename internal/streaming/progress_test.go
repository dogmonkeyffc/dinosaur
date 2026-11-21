package streaming

import (
	"context"
	"io"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"dinosaur/internal/playback"
)

func TestTranscodeReportsSpeed(t *testing.T) {
	ffmpeg, err := exec.LookPath("ffmpeg")
	if err != nil {
		t.Skip("FFmpeg is not installed")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	path := filepath.Join(t.TempDir(), "sample.mkv")
	fixture := exec.CommandContext(ctx, ffmpeg, "-v", "error", "-f", "lavfi", "-i", "color=s=64x48:d=2", "-c:v", "libx264", "-threads", "1", path)
	if output, err := fixture.CombinedOutput(); err != nil {
		t.Fatalf("fixture: %s %v", output, err)
	}
	transcoder := New(ffmpeg)
	defer transcoder.Close()
	stream, err := transcoder.Open(ctx, path, playback.Plan{Mode: "transcode", Audio: -1})
	if err != nil {
		t.Fatal(err)
	}
	defer stream.Close()
	n, err := io.Copy(io.Discard, stream)
	_ = stream.Close()
	if err != nil || n == 0 || stream.TranscodeSpeed() <= 0 {
		t.Fatalf("stream bytes %d, speed %v, error %v, FFmpeg diagnostics %s", n, stream.TranscodeSpeed(), err, stream.Diagnostics())
	}
}

func TestFFmpegProgress(t *testing.T) {
	tail := &stderrTail{}
	p := &ffmpegProgress{diagnostics: tail}
	s := &Stream{progress: p}
	for _, chunk := range []string{"frame=31\nspe", "ed=0.81x\r\nprogress=continue\nspeed=N/A\nspeed=+Infx\nspeed=NaNx\n", "cannot decode stream\n"} {
		if n, err := p.Write([]byte(chunk)); err != nil || n != len(chunk) {
			t.Fatalf("write = %d, %v", n, err)
		}
	}
	if s.TranscodeSpeed() != .81 || tail.String() != "cannot decode stream\n" {
		t.Fatalf("speed %v, diagnostics %q", s.TranscodeSpeed(), tail.String())
	}
	_, _ = p.Write([]byte(strings.Repeat("x", 10000) + "speed=100x\nspeed=1.25x\n"))
	if s.TranscodeSpeed() != 1.25 || cap(p.line) > 16384 {
		t.Fatalf("speed %v, line capacity %d", s.TranscodeSpeed(), cap(p.line))
	}
	if (&Stream{}).TranscodeSpeed() != 0 {
		t.Fatal("missing progress must report no speed")
	}
}
