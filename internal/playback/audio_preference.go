package playback

import (
	"regexp"
	"strings"
)

var originalAudioTitle = regexp.MustCompile(`(?i)(原声|原聲|原音|(^|[^a-z])(original|undubbed)([^a-z]|$))`)
var commentaryAudioTitle = regexp.MustCompile(`(?i)(解说|解說|导评|導評|评论|評論|口述|(^|[^a-z])(commentary|description)([^a-z]|$))`)
var dubbedAudioTitle = regexp.MustCompile(`(?i)(配音|译制|譯製|(^|[^a-z])(dub|dubbed|dubbing)([^a-z]|$))`)

func audioLanguage(value string) string {
	value = strings.ToLower(strings.ReplaceAll(strings.TrimSpace(value), "_", "-"))
	value = strings.Split(value, "-")[0]
	aliases := map[string]string{"eng": "en", "chi": "zh", "zho": "zh", "cmn": "zh", "jpn": "ja", "kor": "ko", "fra": "fr", "fre": "fr", "deu": "de", "ger": "de", "spa": "es", "ita": "it", "por": "pt", "rus": "ru", "hin": "hi", "tha": "th", "vie": "vi", "ara": "ar", "nld": "nl", "dut": "nl", "cn": "yue"}
	if alias := aliases[value]; alias != "" {
		return alias
	}
	if value == "und" || value == "mul" || value == "zxx" || value == "unknown" {
		return ""
	}
	return value
}

func defaultAudioIndex(tracks []Audio) int {
	for _, track := range tracks {
		if track.Default {
			return track.Index
		}
	}
	if len(tracks) > 0 {
		return tracks[0].Index
	}
	return -1
}

// Metadata display language is deliberately excluded: original French/Japanese
// audio must not be replaced merely because the UI or translated title is Chinese.
func PreferredAudio(tracks []Audio, originalLanguage string) int {
	originalLanguage = audioLanguage(originalLanguage)
	best := -100000
	selected := -1
	for _, track := range tracks {
		score := 0
		originalTitle := hasOriginalAudioTitle(track.Title)
		if track.Original {
			score += 10000
		} else if originalTitle {
			score += 8000
		}
		if originalLanguage != "" && trackAudioLanguage(track) == originalLanguage {
			score += 5000
		}
		if track.Default {
			score += 100
		}
		if track.Dubbed || dubbedAudioTitle.MatchString(track.Title) {
			score -= 1000
		}
		if track.Commentary || commentaryAudioTitle.MatchString(track.Title) {
			score -= 30000
		}
		if score > best {
			best = score
			selected = track.Index
		}
	}
	return selected
}

func trackAudioLanguage(track Audio) string {
	language := audioLanguage(track.Language)
	title := strings.ToLower(track.Title)
	if language == "zh" || language == "" {
		if strings.Contains(title, "粤语") || strings.Contains(title, "粵語") || strings.Contains(title, "cantonese") {
			return "yue"
		}
	}
	return language
}

func hasOriginalAudioTitle(title string) bool {
	if !originalAudioTitle.MatchString(title) {
		return false
	}
	lower := strings.ToLower(title)
	for _, negative := range []string{"非原声", "非原聲", "not original", "non-original", "non original", "no original", "without original"} {
		if strings.Contains(lower, negative) {
			return false
		}
	}
	return true
}
