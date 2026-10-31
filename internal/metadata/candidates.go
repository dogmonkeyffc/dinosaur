package metadata

import (
	"sort"
	"strings"
)

// TopCandidates ranks title similarity and release year, preserving provider
// order for ties. It never changes which films are automatically identified.
func TopCandidates(query string, year int, candidates []Candidate) []Candidate {
	result := make([]Candidate, 0, 5)
	seen := make(map[struct {
		provider string
		id       int64
	}]bool)
	for _, candidate := range candidates {
		provider := candidate.Provider
		if provider == "" {
			provider = "tmdb"
		}
		key := struct {
			provider string
			id       int64
		}{provider, candidate.ID}
		if candidate.ID <= 0 || seen[key] {
			continue
		}
		seen[key] = true
		result = append(result, candidate)
	}
	score := func(candidate Candidate) float64 {
		value := max(titleSimilarity(query, candidate.Title), titleSimilarity(query, candidate.OriginalTitle)) * 100
		if year > 0 && candidate.Year > 0 {
			difference := year - candidate.Year
			if difference < 0 {
				difference = -difference
			}
			if difference == 0 {
				value += 25
			} else if difference == 1 {
				value += 10
			}
		}
		return value
	}
	sort.SliceStable(result, func(i, j int) bool { return score(result[i]) > score(result[j]) })
	return result[:min(len(result), 5)]
}

func titleSimilarity(left, right string) float64 {
	left, right = normalizedTitle(left), normalizedTitle(right)
	if left == "" || right == "" {
		return 0
	}
	if left == right {
		return 1
	}
	if strings.Contains(left, right) || strings.Contains(right, left) {
		return float64(min(len([]rune(left)), len([]rune(right)))) / float64(max(len([]rune(left)), len([]rune(right))))
	}
	pairs := func(value string) map[string]int {
		runes := []rune(value)
		result := make(map[string]int)
		for i := 1; i < len(runes); i++ {
			result[string(runes[i-1:i+1])]++
		}
		return result
	}
	a, b := pairs(left), pairs(right)
	intersection, total := 0, 0
	for pair, count := range a {
		intersection += min(count, b[pair])
		total += count
	}
	for _, count := range b {
		total += count
	}
	if total == 0 {
		return 0
	}
	return 2 * float64(intersection) / float64(total)
}
