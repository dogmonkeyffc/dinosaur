package server

import (
	"context"
	"testing"

	"dinosaur/internal/store"
)

func TestStartupInterruptsScansWithoutSchedulingNewJobs(t *testing.T) {
	s, db := fixture(t)
	ctx := context.Background()
	for _, kind := range []string{"movie", "tv"} {
		item, err := db.SaveLibrary(ctx, store.Library{Name: kind, Type: kind, RootPath: t.TempDir()})
		if err != nil {
			t.Fatal(err)
		}
		if kind == "tv" {
			continue
		}
		job, err := db.NewScan(ctx, item.ID)
		if err != nil {
			t.Fatal(err)
		}
		job.Status = "scanning"
		if err := db.SaveScan(ctx, job); err != nil {
			t.Fatal(err)
		}
	}

	if err := s.RecoverInterruptedJobs(ctx); err != nil {
		t.Fatal(err)
	}
	movieJob, err := db.LatestScan(ctx, 1)
	if err != nil || movieJob == nil || movieJob.Status != "interrupted" {
		t.Fatalf("movie scan = %+v, err = %v", movieJob, err)
	}
	tvJob, err := db.LatestScan(ctx, 2)
	if err != nil || tvJob != nil {
		t.Fatalf("unexpected TV scan = %+v, err = %v", tvJob, err)
	}
}
