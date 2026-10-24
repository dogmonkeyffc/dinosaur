package playback

import (
	"errors"
	"math"
	"path/filepath"
	"strings"
)

type Audio struct {
	Index int
	Codec string
}
type Source struct {
	Filename   string
	VideoCodec string
	Duration   float64
	Audio      []Audio
}

type Plan struct {
	Mode  string  `json:"mode"`
	Audio int     `json:"audio"`
	Start float64 `json:"start"`
}

// SelectPlan uses a small portable browser baseline; the client can explicitly
// request compatibility mode if its decoder rejects a nominally supported file.
func SelectPlan(source Source, audio *int, compatible bool, start float64) (Plan, error) {
	if math.IsNaN(start) || math.IsInf(start, 0) || start < 0 || start > 86400 || (source.Duration > 0 && start >= source.Duration) {
		return Plan{}, ErrPosition
	}
	plan := Plan{Mode: "transcode", Audio: -1, Start: start}
	first := -1
	audioCodec := ""
	if len(source.Audio) > 0 {
		first = source.Audio[0].Index
		plan.Audio = first
		audioCodec = source.Audio[0].Codec
	}
	if audio != nil {
		found := false
		for _, track := range source.Audio {
			if track.Index == *audio {
				found = true
				plan.Audio = track.Index
				audioCodec = track.Codec
				break
			}
		}
		if !found {
			return plan, ErrAudio
		}
	}
	if source.VideoCodec == "" {
		return plan, errors.New("video information unavailable")
	}
	if compatible {
		return plan, nil
	}
	extension := strings.ToLower(filepath.Ext(source.Filename))
	defaultAudio := plan.Audio == first
	if defaultAudio && source.VideoCodec == "h264" && (audioCodec == "" || audioCodec == "aac") && (extension == ".mp4" || extension == ".m4v") {
		plan.Mode = "direct"
	} else if defaultAudio && extension == ".webm" && (source.VideoCodec == "vp8" || source.VideoCodec == "vp9") && (audioCodec == "" || audioCodec == "opus" || audioCodec == "vorbis") {
		plan.Mode = "direct"
	} else if source.VideoCodec == "h264" {
		plan.Mode = "remux"
	}
	return plan, nil
}
