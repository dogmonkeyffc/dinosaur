package television

import (
	"errors"
	"strings"
	"testing"
)

func TestValidateWork(t *testing.T) {
	for _, test := range []struct {
		title string
		ids   []int64
	}{
		{" \t\n", []int64{1}}, {"Title", nil}, {"Title", []int64{1, 1}},
		{"Title", []int64{0}}, {"Title", []int64{-1}},
	} {
		_, err := ValidateWork(test.title, test.ids)
		var invalid *WorkError
		if !errors.As(err, &invalid) || invalid.Code != "invalid_work" {
			t.Fatalf("invalid request accepted: %+v %v", test, err)
		}
	}
	title, err := ValidateWork("  作品名 \n", []int64{2, 1})
	if err != nil || title != "作品名" {
		t.Fatalf("title=%q err=%v", title, err)
	}
}

func TestValidateWorkLimits(t *testing.T) {
	ids := make([]int64, 257)
	for i := range ids {
		ids[i] = int64(i + 1)
	}
	if _, err := ValidateWork(" "+strings.Repeat("作", 128)+" ", ids[:256]); err != nil {
		t.Fatalf("boundary should accept 128 Unicode characters and 256 members: %v", err)
	}
	for _, test := range []struct {
		title string
		ids   []int64
	}{
		{strings.Repeat("作", 129), ids[:1]},
		{strings.Repeat("a", 129), ids[:1]},
		{"Title", ids},
	} {
		_, err := ValidateWork(test.title, test.ids)
		var invalid *WorkError
		if !errors.As(err, &invalid) || invalid.Code != "invalid_work" {
			t.Fatalf("oversized request accepted: %v", err)
		}
	}
}

func TestValidateWorkMembership(t *testing.T) {
	for _, test := range []struct {
		library, memberLibrary, work, memberWork int64
		code                                     string
	}{
		{0, 2, 0, 0, ""}, {2, 2, 3, 3, ""}, {2, 2, 3, 0, ""},
		{2, 1, 3, 0, "cross_library"}, {2, 2, 3, 4, "work_conflict"}, {0, 2, 0, 4, "work_conflict"},
	} {
		err := ValidateWorkMember(test.library, test.memberLibrary, test.work, test.memberWork)
		if test.code == "" {
			if err != nil {
				t.Fatal(err)
			}
			continue
		}
		var invalid *WorkError
		if !errors.As(err, &invalid) || invalid.Code != test.code {
			t.Fatalf("case=%+v err=%v", test, err)
		}
	}
}
