package television

import (
	"path/filepath"
	"regexp"
)

var standaloneEpisode = regexp.MustCompile(`(?i)(?:^|[ ._\-\[])e([0-9]{1,4})(?:[ ._\-]?e([0-9]{1,4}))?(?:$|[ ._\-\]\[(])`)
var leadingEpisode = regexp.MustCompile(`^([0-9]{1,3})[ ]*-[ ]*([^0-9 ._-].*)$`)

func numberedEpisodeStem(stem, folder string) string {
	season := "1"
	if parts := numberedSeasonFolder.FindStringSubmatch(filepath.Base(folder)); parts != nil {
		season = parts[1] + parts[2]
	}
	if loc := standaloneEpisode.FindStringSubmatchIndex(stem); loc != nil {
		end := loc[3]
		mark := " S" + season + "E" + stem[loc[2]:end]
		if loc[4] >= 0 {
			end = loc[5]
			mark += "E" + stem[loc[4]:end]
		}
		return stem[:loc[0]] + mark + stem[end:]
	}
	if folder != "." {
		if parts := leadingEpisode.FindStringSubmatch(stem); parts != nil {
			return "S" + season + "E" + parts[1] + " " + parts[2]
		}
	}
	return stem
}
