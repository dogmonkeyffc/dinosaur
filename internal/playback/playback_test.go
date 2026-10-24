package playback

import (
	"math"
	"testing"
)

func TestPlaybackPlan(t *testing.T) {
	for _, test := range []struct{ file, codec, audio, mode string }{
		{"movie.mp4", "h264", "aac", "direct"},
		{"movie.webm", "vp9", "opus", "direct"},
		{"movie.mkv", "h264", "aac", "remux"},
		{"movie.mkv", "hevc", "ac3", "transcode"},
	} {
		source := Source{Filename: test.file, VideoCodec: test.codec, Duration: 600, Audio: []Audio{{Index: 1, Codec: test.audio}, {Index: 2, Codec: "aac"}}}
		plan, err := SelectPlan(source, nil, false, 10)
		if err != nil || plan.Mode != test.mode {
			t.Fatalf("%s: %+v %v", test.file, plan, err)
		}
		compatible, err := SelectPlan(source, nil, true, 10)
		if err != nil || compatible.Mode != "transcode" {
			t.Fatal("compatibility mode not forced")
		}
		invalid := 999
		if _, err := SelectPlan(source, &invalid, false, 0); err == nil {
			t.Fatal("invalid audio accepted")
		}
		for _, position := range []float64{-1, 600, math.NaN(), math.Inf(1)} {
			if _, err := SelectPlan(source, nil, false, position); err == nil {
				t.Fatalf("invalid start accepted: %f", position)
			}
		}
	}
}
func TestProgress(t *testing.T) {
	position := 125.0
	state, err := UpdateProgress(Progress{MediaID: 1}, 1000, &position, nil)
	if err != nil || state.ResumePosition() != position {
		t.Fatalf("resume=%+v %v", state, err)
	}
	nearEnd := 950.0
	state, err = UpdateProgress(state, 1000, &nearEnd, nil)
	if err != nil || !state.Watched || state.ResumePosition() != 0 {
		t.Fatal("completion not detected")
	}
	watched := false
	state, err = UpdateProgress(state, 1000, nil, &watched)
	if err != nil || state.Watched || state.Position != 0 {
		t.Fatal("unwatched did not reset progress")
	}
	for _, invalid := range []float64{-1, math.NaN(), math.Inf(1), 86401} {
		if _, err := UpdateProgress(state, 1000, &invalid, nil); err == nil {
			t.Fatal("invalid progress accepted")
		}
	}
}
