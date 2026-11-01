package television

import (
	"path/filepath"
	"regexp"
	"strings"
)

var chineseEpisode = regexp.MustCompile(`(?:第([0-9]{1,3})季[ ._\-]*)?第([0-9]{1,4})集`)
var numberedSeasonFolder = regexp.MustCompile(`(?i)^(?:(?:season[ ._\-]*|s)([0-9]{1,3})|第([0-9]{1,3})季)$`)

// Normalize only the parsed stem, retaining the original path as series identity.
// A standalone chapter directory does not imply its place in a larger franchise.
func chineseEpisodeStem(stem, folder string) string {
	marks := chineseEpisode.FindAllStringSubmatchIndex(stem, -1)
	if len(marks) != 1 {
		return stem
	}
	mark := marks[0]
	season := "1"
	if mark[2] >= 0 {
		season = stem[mark[2]:mark[3]]
	} else if parts := numberedSeasonFolder.FindStringSubmatch(filepath.Base(folder)); parts != nil {
		season = parts[1] + parts[2]
	}
	episode := stem[mark[4]:mark[5]]
	tail := stem[mark[1]:]
	if end := strings.Index(tail, "]"); end >= 0 {
		tail = tail[:end]
	}
	return stem[:mark[0]] + " S" + season + "E" + episode + " " + tail
}
