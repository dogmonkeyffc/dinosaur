package metadata

func (m *Movie) Matched() bool {
	return m != nil && (m.TMDBID > 0 || m.BangumiID > 0)
}

func (m *Movie) Identity() (string, int64) {
	if m == nil {
		return "tmdb", 0
	}
	if m.BangumiID > 0 {
		return "bangumi", m.BangumiID
	}
	return "tmdb", m.TMDBID
}
