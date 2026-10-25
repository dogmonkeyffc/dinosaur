package scraping

import (
	"context"
	"testing"

	"dinosaur/internal/metadata"
)

func TestReviewCandidatesAreCachedAndClearedOnMatch(t *testing.T) {
	m, db, provider := scrapeFixture(t)
	ctx := context.Background()
	for id := int64(11); id < 18; id++ {
		provider.candidates = append(provider.candidates, metadata.Candidate{ID: id, Title: "Example", Year: 2024})
	}
	job, err := m.Start(ctx, metadata.ScrapeRequest{})
	if err != nil {
		t.Fatal(err)
	}
	waitScrape(t, m, db, job.ID)
	for range 2 {
		result, err := m.Candidates(ctx, 1, "", nil)
		if err != nil || len(result.Candidates) != 5 {
			t.Fatalf("missing cached suggestions: %+v %v", result, err)
		}
	}
	if searches, _ := provider.calls(); searches != 1 {
		t.Fatal("cached suggestions repeated provider search")
	}
	report, err := db.MetadataRecords(ctx, metadata.ReportFilter{Page: 1, PageSize: 20})
	if err != nil || len(report.Items) != 1 || len(report.Items[0].Candidates) != 5 {
		t.Fatalf("admin report lost suggestions: %+v %v", report, err)
	}
	job, err = m.Start(ctx, metadata.ScrapeRequest{MediaID: 1, TMDBID: 10})
	if err != nil {
		t.Fatal(err)
	}
	waitScrape(t, m, db, job.ID)
	// A stale legacy request must not replace a film already matched meanwhile.
	if err := db.RecordScrapeReview(ctx, 1, provider.candidates); err != nil {
		t.Fatal(err)
	}
	item, err := db.MediaByID(ctx, 1)
	if err != nil || item.Metadata.TMDBID != 10 || len(item.Metadata.Candidates) != 0 || item.Metadata.Status != "matched" {
		t.Fatalf("match lost or suggestions survived: %+v %v", item, err)
	}
}

func TestLegacyReviewLoadsCandidatesOnce(t *testing.T) {
	m, db, provider := scrapeFixture(t)
	ctx := context.Background()
	if err := db.RecordScrapeFailure(ctx, 1, "needs_review", "ambiguous"); err != nil {
		t.Fatal(err)
	}
	for range 2 {
		result, err := m.Candidates(ctx, 1, "", nil)
		if err != nil || len(result.Candidates) != 1 {
			t.Fatalf("legacy suggestions not available: %+v %v", result, err)
		}
	}
	if searches, _ := provider.calls(); searches != 1 {
		t.Fatal("legacy suggestions were not persisted")
	}
	if _, err := m.Candidates(ctx, 1, "Different", nil); err != nil {
		t.Fatal(err)
	}
	if searches, _ := provider.calls(); searches != 2 {
		t.Fatal("manual search incorrectly reused original suggestions")
	}
}
