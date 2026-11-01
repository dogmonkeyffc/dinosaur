package streaming

import (
	"encoding/json"
	"os/exec"
	"path/filepath"
	"testing"

	"dinosaur/internal/playback"
)

func TestSelectedQualityProducesExpectedDimensions(t *testing.T) {
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		t.Skip("FFmpeg is not installed")
	}
	if _, err := exec.LookPath("ffprobe"); err != nil {
		t.Skip("FFprobe is not installed")
	}
	for _, test := range []struct {
		quality, input string
		width, height  int
	}{
		{"1080", "1920x1080", 1920, 1080},
		{"720", "1920x1080", 1280, 720},
		{"480", "1920x1080", 854, 480},
		{"original", "1920x1080", 1920, 1080},
		{"1080", "640x360", 640, 360},
		{"720", "720x1280", 406, 720},
	} {
		t.Run(test.quality+"-"+test.input, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "output.mp4")
			filter := videoScale(playback.Plan{Quality: test.quality})
			args := []string{"-v", "error", "-f", "lavfi", "-i", "color=s=" + test.input,
				"-vf", filter, "-frames:v", "1", "-c:v", "libx264", "-threads", "1", path}
			if out, err := exec.Command("ffmpeg", args...).CombinedOutput(); err != nil {
				t.Fatalf("scale failed: %v %s", err, out)
			}
			out, err := exec.Command("ffprobe", "-v", "error", "-show_entries", "stream=width,height", "-of", "json", path).Output()
			var probe struct{ Streams []struct{ Width, Height int } }
			if err != nil || json.Unmarshal(out, &probe) != nil || len(probe.Streams) != 1 {
				t.Fatalf("cannot probe scaled video: %v %s", err, out)
			}
			// FFmpeg versions round the aspect-ratio width to neighboring even pixels.
			width := probe.Streams[0].Width
			if width > test.width || width < test.width-2 || width%2 != 0 || probe.Streams[0].Height != test.height {
				t.Fatalf("unexpected dimensions: %+v", probe.Streams[0])
			}
		})
	}
}
