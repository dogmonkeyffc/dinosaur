package television

import (
	"errors"
	"fmt"
	"io/fs"
	"path"
	"strings"

	"dinosaur/internal/media"
)

var ErrEpisodeAssignment = errors.New("invalid episode assignment")
var ErrEpisodeAssigned = errors.New("episode is already indexed or has a different assignment")

type EpisodeAssignment struct {
	Path         string `json:"path"`
	SeriesFolder string `json:"series_folder"`
	Season       int    `json:"season"`
	Episode      int    `json:"episode"`
}

func (a EpisodeAssignment) Name() (EpisodeName, error) {
	if !publicMediaPath(a.Path) || !publicMediaPath(a.SeriesFolder) ||
		!strings.HasPrefix(a.Path, a.SeriesFolder+"/") || !media.IsVideo(a.Path) ||
		a.Season < 0 || a.Season > 999 || a.Episode < 1 || a.Episode > 9999 {
		return EpisodeName{}, ErrEpisodeAssignment
	}
	title, year := media.ParseFilename(path.Base(a.SeriesFolder) + ".mp4")
	return EpisodeName{
		Key: "folder:" + a.SeriesFolder, Series: title, Year: year,
		Season: a.Season, Episode: a.Episode, End: a.Episode,
		Title: fmt.Sprintf("S%02dE%02d", a.Season, a.Episode),
	}, nil
}

func publicMediaPath(p string) bool {
	return fs.ValidPath(p) && !strings.HasPrefix(p, ".") &&
		!strings.Contains(p, "/.") && !strings.Contains(p, "\\")
}
