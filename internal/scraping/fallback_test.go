package scraping

import (
	"context"
	"errors"
	"testing"

	"dinosaur/internal/metadata"
)

type secondaryProvider struct{ *fakeProvider }

func (p secondaryProvider) Search(ctx context.Context, s metadata.Settings, q string, y int) ([]metadata.Candidate, error) {
	if s.Credential != "" || s.CredentialType != "" {
		return nil, errors.New("TMDB credential reached fallback")
	}
	return p.fakeProvider.Search(ctx, s, q, y)
}
func (p secondaryProvider) Movie(ctx context.Context, s metadata.Settings, id int64) (metadata.Movie, error) {
	if s.Credential != "" || s.CredentialType != "" {
		return metadata.Movie{}, errors.New("TMDB credential reached fallback")
	}
	return p.fakeProvider.Movie(ctx, s, id)
}
func fallbackFixture() *fakeProvider {
	return &fakeProvider{candidates: []metadata.Candidate{{ID: 10, Title: "Example", Year: 2024}}, movie: metadata.Movie{BangumiID: 10, Title: "Bangumi 中文", Year: 2024, Status: "matched", OriginalLanguage: "ja"}}
}

func TestFallbackSelectionPersistenceRefreshAndHistory(t *testing.T) {
	m, db, primary := scrapeFixture(t)
	fallback := fallbackFixture()
	m.WithFallback(secondaryProvider{fallback})
	primary.candidates = nil
	ctx := context.Background()
	job, err := m.Start(ctx, metadata.ScrapeRequest{})
	if err != nil {
		t.Fatal(err)
	}
	if result := waitScrape(t, m, db, job.ID); result.Matched != 1 {
		t.Fatalf("job: %+v", result)
	}
	item, err := db.MediaByID(ctx, 1)
	if err != nil || item.Metadata.BangumiID != 10 || item.Metadata.TMDBID != 0 || item.OriginalLanguage != "ja" {
		t.Fatalf("persisted: %+v %v", item, err)
	}
	page, err := db.MetadataRecords(ctx, metadata.ReportFilter{Page: 1, PageSize: 20})
	if err != nil || page.Summary["matched"] != 1 || page.Items[0].BangumiID != 10 {
		t.Fatalf("records: %+v %v", page, err)
	}
	detail, err := db.ScrapeJobDetail(ctx, job.ID, metadata.ReportFilter{Page: 1, PageSize: 20})
	if err != nil || detail.Items[0].BangumiID != 10 {
		t.Fatalf("history: %+v %v", detail, err)
	}
	targets, err := db.ScrapeTargets(ctx, metadata.ScrapeRequest{})
	if err != nil || len(targets) != 0 {
		t.Fatal("fallback match was queued again")
	}
	before, _ := primary.calls()
	fallback.failure = metadata.ErrProviderUnavailable
	job, err = m.Start(ctx, metadata.ScrapeRequest{MediaID: 1, Refresh: true})
	if err != nil {
		t.Fatal(err)
	}
	waitScrape(t, m, db, job.ID)
	after, _ := primary.calls()
	item, err = db.MediaByID(ctx, 1)
	if err != nil || item.Metadata.Status != "matched" || item.Metadata.BangumiID != 10 || before != after {
		t.Fatalf("refresh did not retain source/info: %+v %v", item, err)
	}
	fallback.failure = nil
	primary.candidates = []metadata.Candidate{{ID: 10, Title: "Example", Year: 2024}}
	job, err = m.Start(ctx, metadata.ScrapeRequest{MediaID: 1, TMDBID: 10})
	if err != nil {
		t.Fatal(err)
	}
	waitScrape(t, m, db, job.ID)
	item, err = db.MediaByID(ctx, 1)
	if err != nil || item.Metadata.TMDBID != 10 || item.Metadata.BangumiID != 0 {
		t.Fatalf("manual source change: %+v %v", item, err)
	}
	detail, err = db.ScrapeJobDetail(ctx, detail.Job.ID, metadata.ReportFilter{Page: 1, PageSize: 20})
	if err != nil || detail.Items[0].BangumiID != 10 || detail.Items[0].TMDBID != 0 {
		t.Fatal("old job source changed after rematch")
	}
}

func TestPrimaryWinsAndFallbackWorksWithoutCredential(t *testing.T) {
	for _, mode := range []string{"primary", "unavailable", "unconfigured", "ambiguous"} {
		t.Run(mode, func(t *testing.T) {
			m, db, primary := scrapeFixture(t)
			fallback := fallbackFixture()
			m.WithFallback(secondaryProvider{fallback})
			ctx := context.Background()
			switch mode {
			case "unavailable":
				primary.failure = metadata.ErrProviderUnavailable
			case "unconfigured":
				if _, err := db.UpdateMetadataSettings(ctx, metadata.SettingsUpdate{}, true); err != nil {
					t.Fatal(err)
				}
			case "ambiguous":
				primary.candidates = append(primary.candidates, metadata.Candidate{ID: 20, Title: "Example", Year: 2024})
			}
			job, err := m.Start(ctx, metadata.ScrapeRequest{})
			if err != nil {
				t.Fatal(err)
			}
			if result := waitScrape(t, m, db, job.ID); result.Matched != 1 {
				t.Fatalf("%+v", result)
			}
			item, err := db.MediaByID(ctx, 1)
			if err != nil {
				t.Fatal(err)
			}
			searches, _ := fallback.calls()
			if mode == "primary" {
				if searches != 0 || item.Metadata.TMDBID != 10 {
					t.Fatal("fallback ran despite primary match")
				}
			} else if searches == 0 || item.Metadata.BangumiID != 10 {
				t.Fatal("fallback was not used")
			}
		})
	}
}

func TestManualCandidatesKeepSourceAndManualBangumiID(t *testing.T) {
	m, db, primary := scrapeFixture(t)
	fallback := fallbackFixture()
	m.WithFallback(secondaryProvider{fallback})
	ctx := context.Background()
	result, err := m.Candidates(ctx, 1, "Example", nil)
	if err != nil || len(result.Candidates) != 2 {
		t.Fatalf("candidates: %+v %v", result, err)
	}
	if result.Candidates[0].ID != 10 || result.Candidates[1].ID != 10 || result.Candidates[0].Provider == result.Candidates[1].Provider {
		t.Fatal("colliding IDs lost their sources")
	}
	job, err := m.Start(ctx, metadata.ScrapeRequest{MediaID: 1, BangumiID: 10})
	if err != nil {
		t.Fatal(err)
	}
	waitScrape(t, m, db, job.ID)
	_, details := primary.calls()
	item, err := db.MediaByID(ctx, 1)
	if err != nil || item.Metadata.BangumiID != 10 || details != 0 {
		t.Fatal("Bangumi candidate used TMDB")
	}
	for _, bad := range []metadata.ScrapeRequest{{BangumiID: 10}, {MediaID: 1, BangumiID: -1}, {MediaID: 1, BangumiID: 10, TMDBID: 10}} {
		if _, err := m.Start(ctx, bad); !errors.Is(err, ErrRequest) {
			t.Fatalf("invalid request accepted: %+v %v", bad, err)
		}
	}
}

type failedDetailProvider struct{ *fakeProvider }

func (p failedDetailProvider) Movie(context.Context, metadata.Settings, int64) (metadata.Movie, error) {
	return metadata.Movie{}, metadata.ErrProviderUnavailable
}

func TestFallbackAfterPrimaryDetailsFail(t *testing.T) {
	m, db, primary := scrapeFixture(t)
	m.provider = failedDetailProvider{primary}
	m.WithFallback(secondaryProvider{fallbackFixture()})
	job, err := m.Start(context.Background(), metadata.ScrapeRequest{})
	if err != nil {
		t.Fatal(err)
	}
	if result := waitScrape(t, m, db, job.ID); result.Matched != 1 {
		t.Fatalf("%+v", result)
	}
	item, err := db.MediaByID(context.Background(), 1)
	if err != nil || item.Metadata.BangumiID != 10 {
		t.Fatal("primary detail failure did not fall back")
	}
}
