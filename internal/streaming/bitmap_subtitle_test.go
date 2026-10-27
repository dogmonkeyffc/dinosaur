package streaming

import (
	"context"
	"image/png"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"dinosaur/internal/playback"
)

// A local PGS sample can be supplied without putting media into the repository.
func TestBitmapSubtitleCompositingAndSeek(t *testing.T) {
	fixture := os.Getenv("DINOSAUR_PGS_FIXTURE")
	if fixture == "" {
		t.Skip("set DINOSAUR_PGS_FIXTURE to a black-video MKV with a PGS cue active from 4.8 to 7.4 seconds")
	}
	for _, start := range []float64{0, 5} {
		t.Run(subtitleClock(start), func(t *testing.T) {
			directory := t.TempDir()
			out := filepath.Join(directory, "burned.mp4")
			frame := filepath.Join(directory, "frame.png")
			index := 1
			args := Arguments(fixture, playback.Plan{Mode: "transcode", Audio: -1, Start: start, Subtitle: &index})
			// Render only the segment needed to inspect the cue, retaining production filter arguments.
			length := "7"
			at := "6"
			if start > 0 {
				length = "2"
				at = "1"
			}
			args = append(args[:len(args)-1], "-t", length, out)
			ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
			defer cancel()
			if data, err := exec.CommandContext(ctx, "ffmpeg", args...).CombinedOutput(); err != nil {
				t.Fatalf("bitmap conversion %v %s", err, data)
			}
			if data, err := exec.CommandContext(ctx, "ffmpeg", "-v", "error", "-ss", at, "-i", out, "-frames:v", "1", frame).CombinedOutput(); err != nil {
				t.Fatalf("frame %v %s", err, data)
			}
			file, err := os.Open(frame)
			if err != nil {
				t.Fatal(err)
			}
			defer file.Close()
			image, err := png.Decode(file)
			if err != nil {
				t.Fatal(err)
			}
			bright := 0
			bottom := 0
			for y := image.Bounds().Min.Y; y < image.Bounds().Max.Y; y++ {
				for x := image.Bounds().Min.X; x < image.Bounds().Max.X; x++ {
					r, g, b, _ := image.At(x, y).RGBA()
					if r > 20000 && g > 20000 && b > 20000 {
						bright++
						if y > image.Bounds().Dy()/2 {
							bottom++
						}
					}
				}
			}
			if bright < 30 || bottom < bright/2 {
				t.Fatalf("PGS cue missing or misplaced after seek: bright=%d bottom=%d", bright, bottom)
			}
		})
	}
}

func TestDelayedBitmapSubtitleDoesNotBlockVideoStartup(t *testing.T) {
	fixture := os.Getenv("DINOSAUR_PGS_DELAYED_FIXTURE")
	if fixture == "" {
		t.Skip("set DINOSAUR_PGS_DELAYED_FIXTURE to a PGS sample whose first cue starts after 30 seconds")
	}
	transcoder := New("ffmpeg")
	defer transcoder.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 7*time.Second)
	defer cancel()
	index := 1
	stream, err := transcoder.Open(ctx, fixture, playback.Plan{Mode: "transcode", Audio: -1, Subtitle: &index})
	if err != nil {
		t.Fatal(err)
	}
	defer stream.Close()
	header := make([]byte, 32)
	if _, err := io.ReadFull(stream, header); err != nil {
		t.Fatalf("video waited for the first subtitle: %v %s", err, stream.Diagnostics())
	}
	stream.Ready()
}
