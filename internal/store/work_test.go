package store

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"dinosaur/internal/media"
	"dinosaur/internal/television"
)

func addWorkSeries(t *testing.T, db *Store, libraryID int64, path string) media.Item {
	t.Helper()
	name, ok := television.ParseEpisode(path)
	if !ok {
		t.Fatalf("invalid episode path %q", path)
	}
	if err := db.SaveEpisode(context.Background(), media.Item{LibraryID: libraryID, RelativePath: path, Title: name.Title, Duration: 100}, name, 1); err != nil {
		t.Fatal(err)
	}
	var id int64
	if err := db.DB.QueryRow(`SELECT id FROM media_items WHERE library_id=? AND relative_path=?`, libraryID, path).Scan(&id); err != nil {
		t.Fatal(err)
	}
	item, err := db.MediaByID(context.Background(), id)
	if err != nil {
		t.Fatal(err)
	}
	return item
}

func requireWorkError(t *testing.T, err error, code string) {
	t.Helper()
	var workErr *television.WorkError
	if !errors.As(err, &workErr) || workErr.Code != code || workErr.Message == "" {
		t.Fatalf("error=%v, want WorkError code=%s with message", err, code)
	}
}

func requireWorkMembers(t *testing.T, db *Store, id int64, title string, members ...int64) {
	t.Helper()
	detail, err := db.WorkByID(context.Background(), id, 0)
	if err != nil || detail.Work.Title != title || len(detail.AllMembers) != len(members) {
		t.Fatalf("detail=%+v error=%v", detail, err)
	}
	for i, member := range members {
		if detail.AllMembers[i].ID != member {
			t.Fatalf("member order=%+v want=%v", detail.AllMembers, members)
		}
	}
}

func TestWorkValidationAndFailedReplacementAreAtomic(t *testing.T) {
	db, lib, episodes := tvFixture(t)
	ctx := context.Background()
	a := episodes[0].SeriesID
	b := addWorkSeries(t, db, lib.ID, "Separate/Beta/S01E01.mp4").SeriesID
	c := addWorkSeries(t, db, lib.ID, "Different/Gamma/S01E01.mp4").SeriesID
	other, err := db.SaveLibrary(ctx, Library{Name: "Other", Type: "tv", RootPath: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	foreign := addWorkSeries(t, db, other.ID, "Foreign/S01E01.mp4").SeriesID
	first, err := db.SaveWork(ctx, 0, "  Ordered work  ", []int64{b, a})
	if err != nil || first.Title != "Ordered work" || first.LibraryID != lib.ID || first.MemberCount != 2 {
		t.Fatalf("work=%+v error=%v", first, err)
	}
	second, err := db.SaveWork(ctx, 0, "Other work", []int64{c})
	if err != nil {
		t.Fatal(err)
	}
	cases := []struct {
		name, title, code string
		ids               []int64
	}{
		{"blank title", " \t", "invalid_work", []int64{a}},
		{"empty members", "Bad", "invalid_work", nil},
		{"duplicate", "Bad", "invalid_work", []int64{a, a}},
		{"zero", "Bad", "invalid_work", []int64{0}},
		{"negative", "Bad", "invalid_work", []int64{-1}},
		{"missing", "Bad", "series_not_found", []int64{a, 999999}},
		{"cross library", "Bad", "cross_library", []int64{a, foreign}},
		{"conflict", "Bad", "work_conflict", []int64{a, c}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := db.SaveWork(ctx, first.ID, tc.title, tc.ids)
			requireWorkError(t, err, tc.code)
			requireWorkMembers(t, db, first.ID, "Ordered work", b, a)
			requireWorkMembers(t, db, second.ID, "Other work", c)
		})
	}
	_, err = db.SaveWork(ctx, 0, "Conflicting create", []int64{foreign, a})
	requireWorkError(t, err, "cross_library")
	works, err := db.Works(ctx, 0, 0)
	if err != nil || len(works) != 2 {
		t.Fatalf("failed create left a partial work: %+v %v", works, err)
	}
	_, err = db.SaveWork(ctx, 999999, "Unknown", []int64{foreign})
	if !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("unknown update=%v", err)
	}
	if _, err := db.WorkByID(ctx, 999999, 0); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("unknown detail=%v", err)
	}
	if err := db.DeleteWork(ctx, 999999); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("unknown delete=%v", err)
	}
	if _, err := db.SaveWork(ctx, first.ID, "Replaced", []int64{a}); err != nil {
		t.Fatal(err)
	}
	requireWorkMembers(t, db, first.ID, "Replaced", a)
	if _, err := db.SaveWork(ctx, 0, "Released member", []int64{b}); err != nil {
		t.Fatalf("full replacement did not release omitted member: %v", err)
	}
}
