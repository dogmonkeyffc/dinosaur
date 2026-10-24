package scanner

import (
	"os"
	"path/filepath"

	"dinosaur/internal/library"
	"dinosaur/internal/media"
)

func (m *Manager) index(item library.Library, relative string, old media.Item, job *library.ScanJob) error {
	path, err := media.ResolveFile(item.RootPath, relative)
	if err != nil {
		job.Errors++
		m.logger.Warn("skip unavailable media file", "library_id", item.ID, "relative_path", relative, "error", err)
		return nil
	}
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	unchanged := old.ID != 0 && old.Size == info.Size() && old.ModifiedNS == info.ModTime().UnixNano() && !old.ProbeError
	if unchanged {
		if err := m.db.TouchMedia(m.ctx, old.ID, job.ID); err != nil {
			return err
		}
		if old.PosterKey != "" {
			if _, err := os.Stat(filepath.Join(m.config.DataDir, "posters", old.PosterKey+".jpg")); err == nil {
				return nil
			}
		}
		key, err := media.BuildPoster(m.ctx, m.config.FFmpegPath, m.config.DataDir, item.RootPath, path, old)
		if err != nil {
			job.Errors++
			return nil
		}
		return m.db.SetMediaPoster(m.ctx, old.ID, key)
	}
	title, year := media.ParseFilename(relative)
	indexed := media.Item{LibraryID: item.ID, RelativePath: relative, Title: title, Year: year, Size: info.Size(), ModifiedNS: info.ModTime().UnixNano()}
	probe, err := media.Probe(m.ctx, m.config.FFprobePath, path)
	if err != nil {
		indexed.ProbeError = true
		job.Errors++
		m.logger.Warn("probe media file failed", "library_id", item.ID, "relative_path", relative, "error", err)
	} else {
		indexed.Duration = probe.Duration
		indexed.Format = probe.Format
		indexed.VideoCodec = probe.VideoCodec
		indexed.Width = probe.Width
		indexed.Height = probe.Height
		indexed.Tracks = probe.Tracks
		key, err := media.BuildPoster(m.ctx, m.config.FFmpegPath, m.config.DataDir, item.RootPath, path, indexed)
		if err != nil {
			job.Errors++
		} else {
			indexed.PosterKey = key
		}
	}
	return m.db.SaveMedia(m.ctx, indexed, job.ID)
}
