package media

import (
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

var releaseYear = regexp.MustCompile(`(?:^|[ ._\-(])((?:19|20)[0-9]{2})(?:$|[ ._\-)])`)
var releaseSuffix = regexp.MustCompile(`(?i)[ ._\-](?:2160p|1080p|720p|480p|bluray|blu-ray|web-dl|webrip|hdtv|dvdrip|x264|x265|h264|h265)(?:[ ._\-]|$)`)

// ParseFilename prefers the title before a release year and keeps a readable
// filename as fallback. It does not invent online metadata or translated names.
func ParseFilename(path string) (string, int) {
	stem := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
	title := stem
	year := 0
	match := releaseYear.FindStringSubmatchIndex(stem)
	if match != nil {
		year, _ = strconv.Atoi(stem[match[2]:match[3]])
		title = stem[:match[0]]
	} else if position := releaseSuffix.FindStringIndex(stem); position != nil {
		title = stem[:position[0]]
	}
	title = strings.Trim(strings.NewReplacer(".", " ", "_", " ").Replace(title), " -[]()")
	title = strings.Join(strings.Fields(title), " ")
	if title == "" {
		title = stem
	}
	return title, year
}

func IsVideo(path string) bool {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".mkv", ".mp4", ".m4v", ".avi", ".mov", ".webm", ".ts", ".m2ts", ".mpg", ".mpeg":
		return true
	default:
		return false
	}
}
