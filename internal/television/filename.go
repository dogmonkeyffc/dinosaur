package television

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"dinosaur/internal/media"
)

var episodeMark = regexp.MustCompile(`(?i)(?:^|[ ._\-])(?:s([0-9]{1,3})e([0-9]{1,4})|([0-9]{1,3})x([0-9]{1,4}))(?:[ ._\-]?e([0-9]{1,4}))?`)
var unsupportedRange = regexp.MustCompile(`(?i)^(?:e[0-9]|-[0-9]|-s[0-9])`)
var seasonFolder = regexp.MustCompile(`(?i)^(?:season[ ._\-]*[0-9]+|s[0-9]+|specials|特别篇|第[0-9一二三四五六七八九十]+季)$`)

// ParseEpisode uses a show directory as identity, falling back to the filename
// for flat libraries. Season zero is reserved for specials. Ambiguous filenames
// without an explicit season and episode are left for the owner to rename.
func ParseEpisode(path string) (EpisodeName, bool) {
	stem := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
	indexes := episodeMark.FindStringSubmatchIndex(stem)
	if indexes == nil {
		return EpisodeName{}, false
	}
	parts := episodeMark.FindStringSubmatch(stem)
	season, episode := parts[1], parts[2]
	if season == "" {
		season, episode = parts[3], parts[4]
	}
	result := EpisodeName{}
	result.Season, _ = strconv.Atoi(season)
	result.Episode, _ = strconv.Atoi(episode)
	result.End = result.Episode
	if parts[5] != "" {
		result.End, _ = strconv.Atoi(parts[5])
	}
	if result.Episode < 1 || result.End < result.Episode {
		return result, false
	}
	// Do not silently interpret a trailing range as a single episode.
	tail := stem[indexes[1]:]
	if unsupportedRange.MatchString(tail) || (len(tail) > 0 && tail[0] >= '0' && tail[0] <= '9') {
		return result, false
	}
	name := strings.Trim(stem[:indexes[0]], " ._-[]()")
	folder := filepath.Dir(path)
	if seasonFolder.MatchString(filepath.Base(folder)) {
		folder = filepath.Dir(folder)
	}
	if folder != "." {
		name = filepath.Base(folder)
		result.Key = "folder:" + filepath.ToSlash(folder)
	}
	if name == "" {
		return result, false
	}
	result.Series, result.Year = media.ParseFilename(name + ".mp4")
	if result.Key == "" {
		result.Key = "title:" + strings.ToLower(result.Series) + fmt.Sprintf(":%d", result.Year)
	}
	result.Title, _ = media.ParseFilename("Episode." + strings.Trim(tail, " ._-[]()") + ".mp4")
	result.Title = strings.TrimSpace(strings.TrimPrefix(result.Title, "Episode"))
	if strings.Trim(tail, " ._-[]()") == "" || result.Title == ".mp4" {
		result.Title = ""
	}
	label := fmt.Sprintf("S%02dE%02d", result.Season, result.Episode)
	if result.End > result.Episode {
		label += fmt.Sprintf("–E%02d", result.End)
	}
	if result.Title == "" {
		result.Title = label
	} else {
		result.Title = label + " · " + result.Title
	}
	return result, true
}
