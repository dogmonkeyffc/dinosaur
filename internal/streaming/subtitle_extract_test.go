package streaming

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"dinosaur/internal/media"
)

func TestEmbeddedSubtitleSegmentsRetainAbsoluteTimeAndInvalidateCache(t *testing.T) {
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		t.Skip("FFmpeg not installed")
	}
	directory := t.TempDir()
	subtitle := filepath.Join(directory, "captions.srt")
	video := filepath.Join(directory, "text.mkv")
	data := "1\n00:00:02,000 --> 00:00:04,000\n中文 early\n\n2\n00:01:10,000 --> 00:01:15,000\n中文 later\n"
	if err := os.WriteFile(subtitle, []byte(data), 0600); err != nil {
		t.Fatal(err)
	}
	for _, codec := range []string{"srt", "ass"} {
		t.Run(codec, func(t *testing.T) {
			path := video + codec + ".mkv"
			cmd := exec.Command("ffmpeg", "-v", "error", "-f", "lavfi", "-i", "color=c=black:s=64x48:r=1:d=150", "-i", subtitle, "-map", "0:v", "-map", "1:s", "-c:v", "libx264", "-preset", "ultrafast", "-threads", "1", "-c:s", codec, path)
			if out, err := cmd.CombinedOutput(); err != nil {
				t.Fatalf("fixture %v %s", err, out)
			}
			extractor := NewSubtitleExtractor("ffmpeg")
			first, err := extractor.Read(context.Background(), path, 1, 0)
			if err != nil || !strings.Contains(string(first), "00:02.000") || !strings.Contains(string(first), "中文 early") {
				t.Fatalf("early segment %v %s", err, first)
			}
			later, err := extractor.Read(context.Background(), path, 1, 70)
			if err != nil || !strings.Contains(string(later), "00:01:10.000 --> 00:01:15.000") || !strings.Contains(string(later), "中文 later") {
				t.Fatalf("late segment %v %s", err, later)
			}
			cancelled, cancel := context.WithCancel(context.Background())
			cancel()
			if _, err := extractor.Read(cancelled, path, 1, 130); err == nil {
				t.Fatal("cancelled extraction succeeded")
			}
			if err := os.Remove(path); err != nil {
				t.Fatal(err)
			}
			if _, err := extractor.Read(context.Background(), path, 1, 70); err == nil {
				t.Fatal("cached subtitles survived removal of the source")
			}
		})
	}
}

func TestSubtitleDiscoveryKeepsExternalIDsAndLabelsEmbeddedTracks(t *testing.T) {
	root := t.TempDir()
	for _, name := range []string{"movie.mkv", "movie.zh.srt"} {
		if err := os.WriteFile(filepath.Join(root, name), nil, 0600); err != nil {
			t.Fatal(err)
		}
	}
	tracks := testSubtitleTracks()
	found := Subtitles(root, "movie.mkv", tracks...)
	if len(found) != 3 || found[0].ID != -3 || found[1].ID != -4 || found[1].Delivery != "burn" || found[2].ID != 0 {
		t.Fatalf("subtitles=%+v", found)
	}
	if !strings.Contains(found[0].Label, "中文") || !strings.Contains(found[1].Label, "PGS") {
		t.Fatal("subtitle format/language missing")
	}
}

func testSubtitleTracks() []media.Track {
	return []media.Track{{Index: 2, Type: "subtitle", Codec: "subrip", Language: "chi"}, {Index: 3, Type: "subtitle", Codec: "hdmv_pgs_subtitle", Language: "eng"}, {Index: 4, Type: "audio", Codec: "aac"}, {Index: 5, Type: "subtitle", Codec: "unknown"}}
}

func TestSubtitleExtractionResourceBounds(t *testing.T) {
	buffer := &subtitleBuffer{}
	if _, err := buffer.Write(make([]byte, 2*1024*1024+1)); err == nil || buffer.Len() != 0 {
		t.Fatal("subtitle output bound not enforced")
	}
	path := filepath.Join(t.TempDir(), "movie.mkv")
	if err := os.WriteFile(path, nil, 0600); err != nil {
		t.Fatal(err)
	}
	extractor := NewSubtitleExtractor("")
	extractor.slots <- struct{}{}
	extractor.slots <- struct{}{}
	if _, err := extractor.Read(context.Background(), path, 1, 0); !errors.Is(err, ErrBusy) {
		t.Fatalf("capacity limit=%v", err)
	}
}
