package store

import (
	"context"
	"fmt"
	"path/filepath"
	"testing"

	"dinosaur/internal/media"
	"dinosaur/internal/metadata"
)

func reportFixture(t *testing.T) (*Store, []media.Item) {
	t.Helper()
	ctx := context.Background()
	db, err := Open(ctx, filepath.Join(t.TempDir(), "reports.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	lib, err := db.SaveLibrary(ctx, Library{Name: "Movies", Type: "movie", RootPath: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	for i := range 4 {
		if err := db.SaveMedia(ctx, media.Item{LibraryID: lib.ID, Title: fmt.Sprintf("Film %d", i), RelativePath: fmt.Sprintf("Film.%d.mp4", i)}, 1); err != nil {
			t.Fatal(err)
		}
	}
	page, err := db.MediaPage(ctx, media.Filter{Page: 1, PageSize: 20})
	if err != nil {
		t.Fatal(err)
	}
	return db, page.Items
}

func TestMetadataReportCountsFiltersAndPagination(t *testing.T) {
	db, items := reportFixture(t)
	ctx := context.Background()
	if err := db.SaveMovieMetadata(ctx, items[0].ID, metadata.Movie{TMDBID: 10, Title: "中文片名", OriginalTitle: "Original", Aliases: "Alias", Status: "matched", Message: "poster_unavailable"}); err != nil {
		t.Fatal(err)
	}
	if err := db.RecordScrapeFailure(ctx, items[1].ID, "needs_review", "ambiguous"); err != nil {
		t.Fatal(err)
	}
	if err := db.RecordScrapeFailure(ctx, items[2].ID, "failed", "provider_unavailable"); err != nil {
		t.Fatal(err)
	}
	filter := metadata.ReportFilter{Page: 1, PageSize: 2}
	result, err := db.MetadataRecords(ctx, filter)
	if err != nil || result.Total != 4 || len(result.Items) != 2 || result.Summary["matched"] != 1 || result.Summary["unmatched"] != 1 || result.Summary["needs_review"] != 1 || result.Summary["failed"] != 1 || result.Summary["warning"] != 1 {
		t.Fatalf("unexpected report: %+v %v", result, err)
	}
	filter.Page = 2
	second, err := db.MetadataRecords(ctx, filter)
	if err != nil || len(second.Items) != 2 || second.Items[0].MediaID == result.Items[0].MediaID {
		t.Fatal("pagination repeated records")
	}
	filter.Page = 1
	for _, status := range []string{"matched", "unmatched", "needs_review", "failed", "warning"} {
		filter.Status = status
		page, err := db.MetadataRecords(ctx, filter)
		if err != nil || page.Total != 1 || page.Summary["all"] != 4 {
			t.Fatalf("invalid status filter %s: %+v %v", status, page, err)
		}
	}
	filter.Status = ""
	for _, query := range []string{"中文", "Original", "Alias", items[0].Filename} {
		filter.Query = query
		page, err := db.MetadataRecords(ctx, filter)
		if err != nil || page.Total != 1 {
			t.Fatalf("invalid search %s: %v", query, err)
		}
	}
	filter.Query, filter.LibraryID = "", 999
	page, err := db.MetadataRecords(ctx, filter)
	if err != nil || page.Total != 0 {
		t.Fatal("library filter leaked other library records")
	}
}

func TestJobSnapshotsSurviveRematchAndMediaRemoval(t *testing.T) {
	db, items := reportFixture(t)
	ctx := context.Background()
	job, err := db.NewScrapeJob(ctx, metadata.ScrapeRequest{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.StartScrapeItems(ctx, job.ID, items); err != nil {
		t.Fatal(err)
	}
	movie := metadata.Movie{TMDBID: 10, Title: "当时片名", Status: "matched"}
	if err := db.SaveMovieMetadata(ctx, items[0].ID, movie); err != nil {
		t.Fatal(err)
	}
	if err := db.SaveScrapeItem(ctx, job.ID, items[0].ID, "matched", ""); err != nil {
		t.Fatal(err)
	}
	job.Status, job.Total, job.Processed, job.Matched, job.Message = "failed", 4, 1, 1, "invalid_credential"
	if err := db.SaveScrapeJob(ctx, job); err != nil {
		t.Fatal(err)
	}
	movie.TMDBID, movie.Title = 20, "后来更正的片名"
	if err := db.SaveMovieMetadata(ctx, items[0].ID, movie); err != nil {
		t.Fatal(err)
	}
	filter := metadata.ReportFilter{Page: 1, PageSize: 2}
	detail, err := db.ScrapeJobDetail(ctx, job.ID, filter)
	if err != nil || detail.Total != 4 || len(detail.Items) != 2 || detail.Items[0].Title != "当时片名" || detail.Items[0].TMDBID != 10 || detail.Items[1].Status != "not_processed" || detail.Items[1].Message != "invalid_credential" {
		t.Fatalf("invalid historical snapshot: %+v %v", detail, err)
	}
	if err := db.DeleteLibrary(ctx, items[0].LibraryID); err != nil {
		t.Fatal(err)
	}
	detail, err = db.ScrapeJobDetail(ctx, job.ID, filter)
	if err != nil || detail.Items[0].MediaID != 0 || detail.Items[0].Filename == "" {
		t.Fatal("deleted source lost historical results")
	}
	filter.Status = "failed"
	history, err := db.ScrapeJobPage(ctx, filter)
	if err != nil || history.Total != 1 || history.Jobs[0].ID != job.ID {
		t.Fatal("job history status filter failed")
	}
	filter.Status = "completed"
	history, err = db.ScrapeJobPage(ctx, filter)
	if err != nil || history.Total != 0 {
		t.Fatal("job history filter ignored")
	}
}
