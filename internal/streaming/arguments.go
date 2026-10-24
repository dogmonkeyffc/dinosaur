package streaming

import (
	"strconv"

	"dinosaur/internal/playback"
)

func Arguments(path string, plan playback.Plan) []string {
	start := strconv.FormatFloat(plan.Start, 'f', 3, 64)
	args := []string{"-hide_banner", "-loglevel", "error", "-nostdin", "-threads", "2", "-ss", start, "-readrate", "1.5", "-readrate_initial_burst", "8", "-protocol_whitelist", "file,pipe", "-i", path}
	if plan.Audio >= 0 {
		// Read audio independently: some MKV tracks start much later than the video.
		// Unthrottled audio discovery plus first_pts=0 supplies the missing silence,
		// allowing the MP4 header to be written before the delayed first audio packet.
		args = append(args, "-threads", "1", "-ss", start, "-protocol_whitelist", "file,pipe", "-i", path)
	}
	args = append(args, "-map", "0:V:0")
	if plan.Audio >= 0 {
		args = append(args, "-map", "1:"+strconv.Itoa(plan.Audio), "-c:a", "aac", "-ac", "2", "-b:a", "192k", "-af", "aresample=async=1:first_pts=0")
	} else {
		args = append(args, "-an")
	}
	args = append(args, "-sn", "-dn", "-map_metadata", "-1", "-map_chapters", "-1", "-filter_threads", "1")
	if plan.Mode == "remux" {
		args = append(args, "-c:v", "copy")
	} else {
		args = append(args, "-c:v", "libx264", "-preset", "veryfast", "-crf", "23", "-pix_fmt", "yuv420p", "-vf", "scale=w='min(1280,iw)':h=-2", "-threads", "2", "-g", "48", "-keyint_min", "48", "-sc_threshold", "0")
	}
	return append(args, "-avoid_negative_ts", "make_zero", "-f", "mp4", "-movflags", "frag_keyframe+empty_moov+default_base_moof", "-frag_duration", "2000000", "pipe:1")
}
