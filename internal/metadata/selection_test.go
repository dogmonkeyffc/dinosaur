package metadata

import "testing"

func TestCandidateSourcePriorityAndExactBangumiMatch(t *testing.T) {
	cases := []struct {
		name   string
		items  []Candidate
		source string
		ok     bool
	}{
		{"TMDB first", []Candidate{{ID: 1, Title: "Example", Year: 2024}, {Provider: "bangumi", ID: 1, Title: "Example", Year: 2024}}, "", true},
		{"ambiguous primary", []Candidate{{ID: 1, Title: "Example", Year: 2024}, {ID: 2, Title: "Example", Year: 2024}, {Provider: "bangumi", ID: 1, Title: "Example", Year: 2024}}, "bangumi", true},
		{"unrelated same year", []Candidate{{Provider: "bangumi", ID: 1, Title: "Different", Year: 2024}}, "", false},
		{"original title", []Candidate{{Provider: "bangumi", ID: 1, Title: "中文", OriginalTitle: "Example", Year: 2023}}, "bangumi", true},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			selected, ok := AutoCandidate("Example", 2024, test.items)
			if ok != test.ok || selected.Provider != test.source {
				t.Fatalf("%+v %v", selected, ok)
			}
		})
	}
	items := TopCandidates("Example", 2024, []Candidate{{ID: 1, Title: "Example"}, {Provider: "bangumi", ID: 1, Title: "Example"}})
	if len(items) != 2 {
		t.Fatal("same ID from different sources was dropped")
	}
	if _, ok := AutoCandidate("Example", 0, items); ok {
		t.Fatal("unknown year auto matched")
	}
}
