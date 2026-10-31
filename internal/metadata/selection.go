package metadata

// AutoCandidate retains source priority and never conflates IDs from two catalogs.
// Bangumi requires an exact title to avoid matching unrelated same-year subjects.
func AutoCandidate(query string, year int, items []Candidate) (Candidate, bool) {
	for _, source := range []string{"tmdb", "bangumi"} {
		group := []Candidate{}
		for _, item := range items {
			name := item.Provider
			if name == "" {
				name = "tmdb"
			}
			if name != source {
				continue
			}
			if source == "bangumi" && normalizedTitle(query) != normalizedTitle(item.Title) && normalizedTitle(query) != normalizedTitle(item.OriginalTitle) {
				continue
			}
			group = append(group, item)
		}
		if id, ok := AutoMatch(query, year, group); ok {
			for _, item := range group {
				if item.ID == id {
					return item, true
				}
			}
		}
	}
	return Candidate{}, false
}
