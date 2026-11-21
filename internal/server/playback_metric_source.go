package server

import (
	"math"
	"net/http"
	"path/filepath"
	"strings"

	"dinosaur/internal/media"
	"dinosaur/internal/playmetrics"
)

func (s *Server) playbackMetricSource(r *http.Request, p playbackMetricWrite) *playmetrics.Source {
	if p.MediaID == nil {
		return nil
	}
	item, err := s.store.MediaByID(r.Context(), *p.MediaID)
	if err != nil {
		return nil
	}
	library, err := s.store.LibraryByID(r.Context(), item.LibraryID)
	if err != nil {
		return nil
	}
	path, err := media.ResolveFile(library.RootPath, item.RelativePath)
	if err != nil || !filepath.IsAbs(path) {
		return nil
	}
	title := item.Title
	if item.SeriesTitle != "" && item.SeriesTitle != title {
		title = item.SeriesTitle + " · " + title
	}
	container := item.Format
	if container == "" {
		container = strings.TrimPrefix(filepath.Ext(path), ".")
	}
	source := &playmetrics.Source{Title: title, Path: path, Width: item.Width, Height: item.Height,
		VideoCodec: item.VideoCodec, Container: container, SizeBytes: item.Size}
	if item.Size > 0 && item.Duration > 0 {
		if rate := float64(item.Size) * 8 / item.Duration; rate > 0 && rate < math.MaxInt64 {
			source.BitrateBPS = int64(math.Round(rate))
		}
	}
	for _, track := range item.Tracks {
		if track.Type == "audio" && p.AudioIndex != nil && track.Index == *p.AudioIndex {
			source.AudioCodec = track.Codec
			break
		}
	}
	return source
}

func playbackMetricReason(p playbackMetricWrite) string {
	switch p.Mode {
	case "direct":
		return "源流直接播放"
	case "remux":
		return "保留视频编码，转换封装或音轨"
	}
	var reasons []string
	if p.BurnSubtitle != nil && *p.BurnSubtitle {
		reasons = append(reasons, "图片字幕烧录")
	}
	switch p.Quality {
	case "1080", "720", "480":
		reasons = append(reasons, "指定 "+p.Quality+"p")
	}
	if p.CompatibilityRequested {
		reasons = append(reasons, "请求兼容模式")
	}
	if len(reasons) == 0 {
		return "源流不满足浏览器直放或封装条件"
	}
	return strings.Join(reasons, " · ")
}
