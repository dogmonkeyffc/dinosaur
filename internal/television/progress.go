package television

import "dinosaur/internal/media"

// ResumeEpisode prefers a partial episode, then the episode after the most
// recently watched one. Specials do not interrupt the main season sequence.
func ResumeEpisode(items []media.Item) *media.Item {
	var partial *media.Item
	latest := -1
	for i := range items {
		item := &items[i]
		if item.Position > 0 && !item.Watched && (partial == nil || item.LastPlayedAt > partial.LastPlayedAt) {
			partial = item
		}
		if item.LastPlayedAt > 0 && (item.Watched || item.Position > 0) && (latest < 0 || item.LastPlayedAt > items[latest].LastPlayedAt) {
			latest = i
		}
	}
	if partial != nil {
		return partial
	}
	if latest >= 0 {
		for i := latest + 1; i < len(items); i++ {
			if !items[i].Watched && items[i].SeasonNumber > 0 {
				return &items[i]
			}
		}
	}
	for i := range items {
		if !items[i].Watched && items[i].SeasonNumber > 0 {
			return &items[i]
		}
	}
	for i := range items {
		if !items[i].Watched {
			return &items[i]
		}
	}
	for i := range items {
		if items[i].SeasonNumber > 0 {
			return &items[i]
		}
	}
	if len(items) > 0 {
		return &items[0]
	}
	return nil
}
