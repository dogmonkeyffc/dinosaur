package streaming

import (
	"fmt"
	"strings"

	"dinosaur/internal/media"
	"dinosaur/internal/playback"
)

func embeddedSubtitles(tracks []media.Track) []Subtitle {
	result := []Subtitle{}
	languages := map[string]string{"chi": "中文", "zho": "中文", "zh": "中文", "eng": "英语", "jpn": "日语", "kor": "韩语", "fre": "法语", "fra": "法语", "und": "未标注语言"}
	formats := map[string]string{"subrip": "SRT", "hdmv_pgs_subtitle": "PGS", "dvd_subtitle": "DVD", "dvb_subtitle": "DVB"}
	for _, track := range tracks {
		delivery := playback.SubtitleDelivery(track.Codec)
		if track.Type != "subtitle" || delivery == "" || track.Index < 0 || track.Index >= 4096 {
			continue
		}
		language := languages[track.Language]
		if language == "" {
			language = track.Language
		}
		if language == "" {
			language = "未标注语言"
		}
		format := formats[track.Codec]
		if format == "" {
			format = strings.ToUpper(track.Codec)
		}
		label := fmt.Sprintf("内嵌 · %s · %s · #%d", language, format, track.Index)
		if strings.TrimSpace(track.Title) != "" {
			label += " · " + track.Title
		}
		if delivery == "burn" {
			label += "（转码）"
		}
		result = append(result, Subtitle{ID: -track.Index - 1, Label: label, Kind: "embedded", Delivery: delivery, Codec: track.Codec, Language: track.Language})
	}
	return result
}
