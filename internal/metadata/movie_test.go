package metadata

import "testing"

func TestAutomaticMatchRequiresUnambiguousYear(t *testing.T) {
	for _, test := range []struct {
		name       string
		query      string
		year       int
		candidates []Candidate
		want       int64
	}{
		{"translated unique year", "The Chorus", 2004, []Candidate{{ID: 1, Title: "放牛班的春天", OriginalTitle: "Les Choristes", Year: 2004}}, 1},
		{"ambiguous", "Example", 2004, []Candidate{{ID: 1, Year: 2004}, {ID: 2, Year: 2004}}, 0},
		{"missing year", "Example", 0, []Candidate{{ID: 1, Title: "Example", Year: 2004}}, 0},
		{"wrong year", "Example", 2004, []Candidate{{ID: 1, Title: "Example", Year: 2010}}, 0},
		{"near year exact title", "The-Chorus", 2005, []Candidate{{ID: 1, Title: "The Chorus", Year: 2004}}, 1},
		{"near year unrelated", "Example", 2005, []Candidate{{ID: 1, Title: "Other", Year: 2004}}, 0},
		{"none", "Example", 2005, nil, 0},
	} {
		t.Run(test.name, func(t *testing.T) {
			id, matched := AutoMatch(test.query, test.year, test.candidates)
			if id != test.want || matched != (test.want > 0) {
				t.Fatalf("id=%d matched=%v", id, matched)
			}
		})
	}
}
