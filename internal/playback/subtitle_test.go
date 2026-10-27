package playback

import (
	"testing"

	"dinosaur/internal/media"
)

func TestBitmapSubtitlePlanRejectsOtherStreamsAndOriginalBypass(t *testing.T) {
	tracks := []media.Track{{Index: 0, Type: "video", Codec: "h264"}, {Index: 1, Type: "audio", Codec: "aac"}, {Index: 2, Type: "subtitle", Codec: "subrip"}, {Index: 3, Type: "subtitle", Codec: "hdmv_pgs_subtitle"}}
	plan, err := WithSubtitle(Plan{Mode: "direct", Audio: 1, Start: 125}, -4, tracks)
	if err != nil || plan.Mode != "transcode" || plan.Subtitle == nil || *plan.Subtitle != 3 || plan.Start != 125 || plan.Audio != 1 {
		t.Fatalf("bitmap plan=%+v %v", plan, err)
	}
	for _, id := range []int{0, -1, -2, -3, -999999} {
		if _, err := WithSubtitle(plan, id, tracks); err == nil {
			t.Fatalf("invalid subtitle %d accepted", id)
		}
	}
}
