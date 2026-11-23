package history

import (
	"testing"
	"time"
)

func TestValidationAndShanghaiDate(t *testing.T) {
	e := Entry{Title: "Example", Source: "douban", SourceID: "123", WatchedOn: "2024-02-29", Rating: 5}
	if err := e.Validate(); err != nil {
		t.Fatal(err)
	}
	for _, day := range []string{"2023-02-29", "2024-2-1", "invalid", "1899-01-01"} {
		bad := e
		bad.WatchedOn = day
		if bad.Validate() == nil {
			t.Fatalf("accepted date %s", day)
		}
	}
	for _, id := range []string{"../secret", "-1", "https://example.com", ""} {
		bad := e
		bad.SourceID = id
		if bad.Validate() == nil {
			t.Fatal("accepted invalid source ID")
		}
	}
	e.Rating = 6
	if e.Validate() == nil {
		t.Fatal("accepted rating 6")
	}
	if Day(time.Date(2026, 11, 23, 17, 0, 0, 0, time.UTC)) != "2026-11-24" {
		t.Fatal("wrong viewing day")
	}
}
