package store

import (
	"context"
	"testing"

	"dinosaur/internal/media"
	"dinosaur/internal/television"
)

func catalogUser(t *testing.T, db *Store, name string) int64 {
	t.Helper()
	u, err := db.CreateUser(context.Background(), name, "hash", "user", false)
	if err != nil {
		t.Fatal(err)
	}
	return u.ID
}

func catalogProgress(t *testing.T, db *Store, user, id, at int64, position float64, watched bool) {
	t.Helper()
	_, err := db.DB.Exec(`INSERT INTO playback_progress(user_id,media_id,updated_at,position,watched) VALUES(?,?,?,?,?)`, user, id, at, position, watched)
	if err != nil {
		t.Fatal(err)
	}
}

func TestMovieCatalogRecentOrderIsPersonalAndPaginated(t *testing.T) {
	db, items := reportFixture(t)
	alice, bob := catalogUser(t, db, "alice"), catalogUser(t, db, "bob")
	catalogProgress(t, db, alice, items[2].ID, 700, 100, true)
	catalogProgress(t, db, alice, items[0].ID, 600, 20, false)
	catalogProgress(t, db, alice, items[3].ID, 900, 0, false)
	catalogProgress(t, db, bob, items[1].ID, 1000, 30, false)
	filter := media.Filter{ViewerID: alice, Page: 1, PageSize: 2, LibraryID: items[0].LibraryID, Query: "Film"}
	for page, expected := range [][]int64{{items[2].ID, items[0].ID}, {items[1].ID, items[3].ID}} {
		filter.Page = page + 1
		result, err := db.MediaPage(context.Background(), filter)
		if err != nil || result.Total != 4 || len(result.Items) != 2 {
			t.Fatalf("page: %+v %v", result, err)
		}
		for i, id := range expected {
			if result.Items[i].ID != id {
				t.Fatalf("page %d: wrong recent order: %+v", page+1, result.Items)
			}
		}
	}
	filter.ViewerID, filter.Page = bob, 1
	result, err := db.MediaPage(context.Background(), filter)
	if err != nil || result.Items[0].ID != items[1].ID {
		t.Fatalf("viewer history leaked: %+v %v", result, err)
	}
}

func TestSeriesCatalogUsesLatestVisibleEpisodePerViewer(t *testing.T) {
	db, lib, episodes := tvFixture(t)
	ctx := context.Background()
	alice, bob := catalogUser(t, db, "alice"), catalogUser(t, db, "bob")
	for _, name := range []string{"Zulu", "Alpha", "Bravo"} {
		path := name + "/S01E01.mp4"
		episode, _ := television.ParseEpisode(path)
		if err := db.SaveEpisode(ctx, media.Item{LibraryID: lib.ID, RelativePath: path, Title: episode.Title}, episode, 1); err != nil {
			t.Fatal(err)
		}
	}
	var zulu, alpha, bravo int64
	for name, id := range map[string]*int64{"Zulu": &zulu, "Alpha": &alpha, "Bravo": &bravo} {
		if err := db.DB.QueryRow(`SELECT m.id FROM media_items m WHERE relative_path=?`, name+"/S01E01.mp4").Scan(id); err != nil {
			t.Fatal(err)
		}
	}
	catalogProgress(t, db, alice, episodes[0].ID, 600, 20, false)
	catalogProgress(t, db, alice, episodes[1].ID, 1300, 100, true)
	catalogProgress(t, db, alice, zulu, 1100, 30, false)
	catalogProgress(t, db, alice, bravo, 1500, 0, false)
	catalogProgress(t, db, bob, alpha, 2000, 40, false)
	if _, err := db.DB.Exec(`UPDATE media_items SET missing=1 WHERE id=?`, episodes[1].ID); err != nil {
		t.Fatal(err)
	}
	filter := media.Filter{LibraryID: lib.ID, Page: 1, PageSize: 2}
	for page, expected := range [][]string{{"Zulu", "Example"}, {"Alpha", "Bravo"}} {
		filter.Page = page + 1
		result, err := db.SeriesPage(ctx, filter, alice)
		if err != nil || result.Total != 4 || len(result.Items) != 2 {
			t.Fatalf("page: %+v %v", result, err)
		}
		for i, name := range expected {
			if result.Items[i].Title != name {
				t.Fatalf("page %d: wrong recent order: %+v", page+1, result.Items)
			}
		}
	}
	filter.Page = 1
	result, err := db.SeriesPage(ctx, filter, bob)
	if err != nil || result.Items[0].Title != "Alpha" {
		t.Fatalf("viewer history leaked: %+v %v", result, err)
	}
}
