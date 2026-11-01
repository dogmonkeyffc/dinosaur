package playback

import "errors"

// WithQuality keeps automatic playback unchanged; an explicit size requires
// video encoding, while original preserves the source dimensions if encoding.
func WithQuality(plan Plan, quality string) (Plan, error) {
	if quality == "" {
		quality = "auto"
	}
	switch quality {
	case "auto", "original":
	case "1080", "720", "480":
		plan.Mode = "transcode"
	default:
		return plan, errors.New("invalid playback quality")
	}
	plan.Quality = quality
	return plan, nil
}
