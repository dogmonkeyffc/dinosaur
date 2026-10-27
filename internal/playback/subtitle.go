package playback

import (
	"errors"

	"dinosaur/internal/media"
)

var ErrSubtitle = errors.New("invalid or unsupported subtitle track")

func SubtitleDelivery(codec string) string {
	switch codec {
	case "subrip", "srt", "ass", "ssa", "webvtt", "mov_text", "text":
		return "text"
	case "hdmv_pgs_subtitle", "dvd_subtitle", "dvb_subtitle":
		return "burn"
	}
	return ""
}

// Bitmap subtitles require decoding and compositing even when original playback was requested.
func WithSubtitle(plan Plan, id int, tracks []media.Track) (Plan, error) {
	if id >= 0 || id < -4096 {
		return plan, ErrSubtitle
	}
	index := -id - 1
	for _, track := range tracks {
		if track.Index == index && track.Type == "subtitle" && SubtitleDelivery(track.Codec) == "burn" {
			plan.Mode = "transcode"
			plan.Subtitle = &index
			return plan, nil
		}
	}
	return plan, ErrSubtitle
}
