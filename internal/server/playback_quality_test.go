package server

import (
	"encoding/json"
	"net/http/httptest"
	"net/url"
	"testing"

	"dinosaur/internal/media"
	"dinosaur/internal/playback"
)

func TestPlaybackQualitySurvivesStreamURLAndSubtitleSelection(t *testing.T) {
	item := media.Item{Filename: "Movie.mp4", VideoCodec: "h264", Duration: 600,
		Tracks: []media.Track{{Index: 1, Type: "audio", Codec: "aac"}, {Index: 2, Type: "subtitle", Codec: "hdmv_pgs_subtitle"}}}
	for _, quality := range []string{"auto", "original", "1080", "720", "480"} {
		for _, subtitle := range []string{"", "&subtitle=-3"} {
			r := httptest.NewRequest("GET", "/?quality="+quality+subtitle, nil)
			plan, err := requestPlan(r, item)
			if err != nil || plan.Quality != quality {
				t.Fatalf("quality %s: %+v %v", quality, plan, err)
			}
			if (quality != "auto" && quality != "original" || subtitle != "") && plan.Mode != "transcode" {
				t.Fatal("selected resolution or bitmap subtitle must transcode")
			}
			stream := httptest.NewRequest("GET", streamURL(1, plan, false)+"&start=125", nil)
			again, err := requestPlan(stream, item)
			if err != nil || again.Quality != quality || again.Mode != plan.Mode || again.Start != 125 || again.Audio != plan.Audio {
				t.Fatalf("stream plan lost selection: %+v %v", again, err)
			}
			if (again.Subtitle == nil) != (plan.Subtitle == nil) {
				t.Fatal("stream lost subtitle selection")
			}
		}
	}
}

func TestPlaybackQualityValidationOnInfoAndStream(t *testing.T) {
	s, _, admin, _ := playbackFixture(t)
	for _, quality := range []string{"999999", "-1", "4k", "720;evil"} {
		for _, path := range []string{"/api/playback/1", "/api/playback/1/stream"} {
			status(t, request(t, s, "GET", path+"?quality="+url.QueryEscape(quality), nil, &admin, nil), 400)
		}
	}
	w := request(t, s, "GET", "/api/playback/1?quality=480", nil, &admin, nil)
	status(t, w, 200)
	var info struct {
		Plan playback.Plan `json:"plan"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &info); err != nil || info.Plan.Quality != "480" || info.Plan.Mode != "transcode" {
		t.Fatalf("quality missing from playback info: %s", w.Body.String())
	}
}
