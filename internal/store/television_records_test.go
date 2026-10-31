package store

import (
	"context"
	"fmt"
	"testing"

	"dinosaur/internal/media"
	"dinosaur/internal/metadata"
	"dinosaur/internal/television"
)

func TestMetadataRecordsIncludeSeriesOnceAndFilterBothLibraries(t *testing.T) {
	db, movies := reportFixture(t)
	ctx := context.Background()
	lib, err := db.SaveLibrary(ctx, Library{Name: "电视剧", Type: "tv", RootPath: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	paths := []string{"Show_A (2024)/Show_A.S01E01.mkv", "Show_A (2024)/Show_A.S01E02.OnlySecond.mkv", "ShowB (2024)/ShowB.S01E01.mkv", "Hidden (2024)/Hidden.S01E01.mkv"}
	for _, path := range paths {
		name, ok := television.ParseEpisode(path)
		if !ok {
			t.Fatalf("invalid fixture episode: %s", path)
		}
		if err := db.SaveEpisode(ctx, media.Item{LibraryID: lib.ID, RelativePath: path, Title: name.Title}, name, 1); err != nil {
			t.Fatal(err)
		}
	}
	repo := TelevisionRepository{Store: db}
	if err := repo.SaveMovieMetadata(ctx, 1, metadata.Movie{BangumiID: 777, Title: "中文剧名", OriginalTitle: "Original Show", Aliases: "Alias Show", Status: "matched", Message: "poster_unavailable", Language: "zh-CN"}); err != nil {
		t.Fatal(err)
	}
	if err := repo.RecordScrapeReview(ctx, 2, []metadata.Candidate{{Provider: "bangumi", ID: 888, Title: "ShowB", Year: 2024}}); err != nil {
		t.Fatal(err)
	}
	if _, err := db.DB.Exec(`UPDATE media_items SET missing=1 WHERE relative_path=?`, paths[3]); err != nil {
		t.Fatal(err)
	}
	if _, err := db.DB.Exec(`UPDATE media_items SET relative_path=? WHERE relative_path=?`, `ShowB (2024)/literal\%_part.mkv`, paths[2]); err != nil {
		t.Fatal(err)
	}
	all := metadata.ReportFilter{Page: 1, PageSize: 20}
	page, err := db.MetadataRecords(ctx, all)
	if err != nil || page.Total != 6 || page.Summary["matched"] != 1 || page.Summary["needs_review"] != 1 || page.Summary["unmatched"] != 4 || page.Summary["warning"] != 1 {
		t.Fatalf("mixed report: %+v %v", page, err)
	}
	movieFilter := all
	movieFilter.LibraryID = movies[0].LibraryID
	moviePage, err := db.MetadataRecords(ctx, movieFilter)
	if err != nil || moviePage.Total != 4 {
		t.Fatalf("movie library: %+v %v", moviePage, err)
	}
	tvFilter := all
	tvFilter.LibraryID = lib.ID
	tvPage, err := db.MetadataRecords(ctx, tvFilter)
	if err != nil || tvPage.Total != 2 {
		t.Fatalf("TV library: %+v %v", tvPage, err)
	}
	for _, item := range tvPage.Items {
		if item.MediaID != 0 || item.SeriesID <= 0 || item.LibraryID != lib.ID {
			t.Fatalf("episode ID used as metadata target: %+v", item)
		}
		if item.SeriesID == 1 && (item.BangumiID != 777 || item.TMDBID != 0 || item.Language != "zh-CN") {
			t.Fatal("series provenance missing")
		}
		if item.SeriesID == 2 && (len(item.Candidates) != 1 || item.Candidates[0].Provider != "bangumi") {
			t.Fatal("series candidates missing")
		}
	}
	for _, status := range []string{"matched", "needs_review", "warning"} {
		f := tvFilter
		f.Status = status
		result, err := db.MetadataRecords(ctx, f)
		if err != nil || result.Total != 1 || result.Summary["all"] != 2 {
			t.Fatalf("status %s: %+v %v", status, result, err)
		}
	}
	for _, query := range []string{"中文剧名", "Original Show", "Alias Show", "Show_A", "OnlySecond", `\%_`} {
		f := all
		f.Query = query
		result, err := db.MetadataRecords(ctx, f)
		if err != nil || result.Total != 1 || result.Items[0].SeriesID == 0 {
			t.Fatalf("query %q: %+v %v", query, result, err)
		}
	}
	seen := map[string]bool{}
	for number := 1; number <= 6; number++ {
		f := all
		f.Page, f.PageSize = number, 1
		result, err := db.MetadataRecords(ctx, f)
		if err != nil || len(result.Items) != 1 {
			t.Fatalf("page %d: %+v %v", number, result, err)
		}
		key := fmt.Sprintf("%d:%d", result.Items[0].MediaID, result.Items[0].SeriesID)
		if seen[key] {
			t.Fatal("overlapping movie/series IDs repeated across pages")
		}
		seen[key] = true
	}
}
