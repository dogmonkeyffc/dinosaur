package store

import (
	"context"
	"testing"
)

func TestWorkWriteFailureRollsBackTitleMembershipAndCreation(t *testing.T) {
	db, lib, episodes := tvFixture(t)
	ctx := context.Background()
	a := episodes[0].SeriesID
	b := addWorkSeries(t, db, lib.ID, "Beta/S01E01.mp4").SeriesID
	c := addWorkSeries(t, db, lib.ID, "Gamma/S01E01.mp4").SeriesID
	work, err := db.SaveWork(ctx, 0, "Original", []int64{b, a})
	if err != nil {
		t.Fatal(err)
	}
	// Force a failure after the title update, membership deletion, and first insert.
	if _, err := db.DB.Exec(`CREATE TRIGGER reject_second_work_member BEFORE INSERT ON work_members WHEN NEW.position=1 BEGIN SELECT RAISE(ABORT,'injected write failure'); END`); err != nil {
		t.Fatal(err)
	}
	if _, err := db.SaveWork(ctx, work.ID, "Must rollback", []int64{a, c}); err == nil {
		t.Fatal("injected replacement failure was ignored")
	}
	requireWorkMembers(t, db, work.ID, "Original", b, a)
	if _, err := db.SaveWork(ctx, 0, "Must not exist", []int64{c, c + 100000}); err == nil {
		t.Fatal("invalid create unexpectedly succeeded")
	}
	// Release both existing members so creation reaches the trigger too.
	if err := db.DeleteWork(ctx, work.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := db.SaveWork(ctx, 0, "Must rollback creation", []int64{c, a}); err == nil {
		t.Fatal("injected creation failure was ignored")
	}
	works, err := db.Works(ctx, lib.ID, 0)
	if err != nil || len(works) != 0 {
		t.Fatalf("failed creation left work: %+v %v", works, err)
	}
	var members int
	if err := db.DB.QueryRow(`SELECT count(*) FROM work_members`).Scan(&members); err != nil || members != 0 {
		t.Fatalf("failed creation left memberships=%d error=%v", members, err)
	}
	if _, err := db.DB.Exec(`DROP TRIGGER reject_second_work_member`); err != nil {
		t.Fatal(err)
	}
	if _, err := db.SaveWork(ctx, 0, "Retry", []int64{c, a}); err != nil {
		t.Fatalf("rollback did not release memberships: %v", err)
	}
}
