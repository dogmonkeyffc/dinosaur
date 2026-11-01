package streaming

import "dinosaur/internal/playback"

func videoScale(plan playback.Plan) string {
	width, height := "1280", "720"
	switch plan.Quality {
	case "original":
		return "scale=w='trunc(iw/2)*2':h='trunc(ih/2)*2'"
	case "1080":
		width, height = "1920", "1080"
	case "480":
		width, height = "854", "480"
	}
	return "scale=w='min(" + width + ",iw)':h='min(" + height + ",ih)':force_original_aspect_ratio=decrease:force_divisible_by=2"
}
