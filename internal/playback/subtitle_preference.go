package playback

import (
	"regexp"
	"strings"
)

// SubtitleOption contains only the metadata needed to choose an initial caption track.
type SubtitleOption struct {
	ID                        int
	Language, Title, Delivery string
}

var subtitleLanguageToken = regexp.MustCompile(`[a-z]+`)
var simplifiedSubtitle = regexp.MustCompile(`(?i)(简体|简中|简英|簡體|簡中|簡英|(^|[^a-z0-9])(chs|hans|zh[-_](cn|sg))([^a-z0-9]|$))`)
var partialSubtitle = regexp.MustCompile(`(?i)(强制|強制|导评|導評|解说|解說|(^|[^a-z0-9])(forced|commentary|signs)([^a-z0-9]|$))`)

func subtitleLanguage(language, title string) string {
	language = strings.ToLower(strings.ReplaceAll(strings.TrimSpace(language), "_", "-"))
	switch strings.Split(language, "-")[0] {
	case "zh", "chi", "zho", "cmn", "yue", "chs", "cht":
		return "zh"
	case "en", "eng":
		// A bilingual Chinese track can be tagged as English by its author.
		if !chineseSubtitleTitle(title) {
			return "en"
		}
	}
	if chineseSubtitleTitle(title) {
		return "zh"
	}
	for _, match := range subtitleLanguageToken.FindAllString(strings.ToLower(title), -1) {
		value := match
		if value == "en" || value == "eng" || value == "english" {
			return "en"
		}
	}
	return ""
}

func chineseSubtitleTitle(title string) bool {
	lowered := strings.ToLower(title)
	for _, negative := range []string{"无中文", "無中文", "非中文", "no chinese", "without chinese", "not chinese"} {
		if strings.Contains(lowered, negative) {
			return false
		}
	}
	for _, word := range []string{"中文", "中字", "简体", "簡體", "繁体", "繁體", "简中", "簡中", "繁中", "中英", "英中", "简英", "簡英", "繁英", "简繁", "簡繁", "繁简", "繁簡"} {
		if strings.Contains(title, word) {
			return true
		}
	}
	for _, match := range subtitleLanguageToken.FindAllString(strings.ToLower(title), -1) {
		switch match {
		case "zh", "chi", "zho", "cmn", "yue", "chs", "cht", "chinese", "mandarin", "cantonese":
			return true
		}
	}
	return false
}

// PreferredSubtitle leaves unmarked or other-language tracks off instead of guessing.
// Text captions avoid transcoding; full captions outrank forced/commentary tracks.
func PreferredSubtitle(options []SubtitleOption) *int {
	best := 0
	var selected *int
	for _, option := range options {
		score := 0
		switch subtitleLanguage(option.Language, option.Title) {
		case "zh":
			score = 10000
		case "en":
			score = 1000
		default:
			continue
		}
		if option.Delivery != "text" && option.Delivery != "burn" {
			continue
		}
		if partialSubtitle.MatchString(option.Title) {
			score -= 500
		}
		if option.Delivery == "text" {
			score += 100
		}
		if simplifiedSubtitle.MatchString(option.Title) || strings.Contains(strings.ToLower(option.Language), "hans") || strings.EqualFold(option.Language, "zh-cn") {
			score += 20
		}
		if score > best {
			best = score
			id := option.ID
			selected = &id
		}
	}
	return selected
}
