package media

import (
	"context"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestProbeAudioDispositions(t *testing.T) {
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		t.Skip("FFmpeg not installed")
	}
	path := filepath.Join(t.TempDir(), "audio.mkv")
	args := []string{"-v", "error", "-f", "lavfi", "-i", "color=s=64x48:d=1", "-f", "lavfi", "-i", "sine=frequency=440:duration=1", "-map", "0:v", "-map", "1:a", "-map", "1:a", "-c:v", "libx264", "-preset", "ultrafast", "-threads", "1", "-c:a", "aac", "-metadata:s:a:0", "language=chi", "-metadata:s:a:1", "language=jpn", "-disposition:a:0", "default+dub", "-disposition:a:1", "original", path}
	if data, err := exec.Command("ffmpeg", args...).CombinedOutput(); err != nil {
		t.Fatalf("fixture: %v %s", err, data)
	}
	result, err := Probe(context.Background(), "ffprobe", path)
	if err != nil || len(result.Tracks) != 2 {
		t.Fatalf("probe: %+v %v", result, err)
	}
	dub, original := result.Tracks[0], result.Tracks[1]
	if dub.Default == nil || !*dub.Default || !dub.Dubbed || original.Default == nil || *original.Default || !original.Original || original.Language != "jpn" {
		t.Fatalf("disposition lost: %+v %+v", dub, original)
	}
}
