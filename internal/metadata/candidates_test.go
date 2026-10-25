package metadata

import "testing"

func TestTopCandidatesRanksDeduplicatesAndLimits(t *testing.T) {
	candidates := []Candidate{
		{ID: 1, Title: "Unrelated", Year: 2024},
		{ID: 2, Title: "中文译名", OriginalTitle: "Example", Year: 2024},
		{ID: 3, Title: "Example", Year: 2023},
		{ID: 4, Title: "Example", Year: 2000},
		{ID: 5, Title: "Example Extended", Year: 2024},
		{ID: 6, Title: "Other", Year: 2000},
		{ID: 7, Title: "Other", Year: 2000},
		{ID: 2, Title: "duplicate"}, {ID: 0, Title: "Example"},
	}
	got := TopCandidates("Example", 2024, candidates)
	want := []int64{2, 3, 4, 5, 1}
	if len(got) != len(want) {
		t.Fatalf("expected five suggestions: %+v", got)
	}
	for i, id := range want {
		if got[i].ID != id {
			t.Fatalf("rank %d: got %d, want %d", i, got[i].ID, id)
		}
	}
	if candidates[0].ID != 1 {
		t.Fatal("ranking mutated provider results")
	}
	if result := TopCandidates("", 0, nil); len(result) != 0 {
		t.Fatal("empty search produced candidates")
	}
}

func TestCandidateSimilaritySupportsChineseAndStableTies(t *testing.T) {
	result := TopCandidates("春天的故事", 0, []Candidate{
		{ID: 1, Title: "秋天的故事"}, {ID: 2, Title: "春天的故事"},
		{ID: 3, Title: "秋天的故事"},
	})
	if result[0].ID != 2 || result[1].ID != 1 || result[2].ID != 3 {
		t.Fatalf("unexpected Chinese ranking or unstable ties: %+v", result)
	}
}
