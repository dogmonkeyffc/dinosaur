package television

import (
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"dinosaur/internal/media"
)

var seasonPack = regexp.MustCompile(`(?i)(?:s[0-9]{1,3}[ ._]*-[ ._]*s[0-9]{1,3}|全[0-9一二三四五六七八九十]+季)`)
var seasonInTitle = regexp.MustCompile(`(?i)(?:[ ._\-]*s([0-9]{1,3})(?:[ ._\-]|$)|[ ._\-]*第([0-9]{1,3})季)`)

// collectionFolder recognizes named season folders under a show directory.
// A matching show prefix prevents a shared parent such as "Comedy" from grouping unrelated shows.
func collectionFolder(folder string, season int) (string, string, int, bool) {
	parent := filepath.Dir(folder)
	if parent == "." {
		return "", "", 0, false
	}
	child := filepath.Base(folder)
	mark := seasonInTitle.FindStringSubmatchIndex(child)
	if mark == nil {
		return "", "", 0, false
	}
	parts := seasonInTitle.FindStringSubmatch(child)
	number, err := strconv.Atoi(parts[1] + parts[2])
	if err != nil || number != season {
		return "", "", 0, false
	}
	name := filepath.Base(parent)
	_, year := media.ParseFilename(name + ".mp4")
	if marker := seasonPack.FindStringIndex(name); marker != nil {
		name = name[:marker[0]]
	}
	title, _ := media.ParseFilename(name + ".mp4")
	childTitle, _ := media.ParseFilename(child[:mark[0]] + ".mp4")
	if title == ".mp4" || !strings.EqualFold(title, childTitle) {
		return "", "", 0, false
	}
	return parent, title, year, true
}
