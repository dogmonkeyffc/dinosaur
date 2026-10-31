package store

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"

	"dinosaur/internal/agent"
	"dinosaur/internal/auth"
	"dinosaur/internal/media"
	"dinosaur/internal/metadata"
	"dinosaur/internal/television"
)

func agentStoreFixture(t *testing.T) (*Store, agent.Key, int64) {
	t.Helper()
	db, _ := reportFixture(t)
	ctx := context.Background()
	user, err := db.CreateUser(ctx, "agent-admin", "hash", "admin", true)
	if err != nil {
		t.Fatal(err)
	}
	tv, err := db.SaveLibrary(ctx, Library{Name: "TV", Type: "tv", RootPath: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	path := "Show (2024)/Show.S01E01.mkv"
	name, _ := television.ParseEpisode(path)
	if err := db.SaveEpisode(ctx, media.Item{LibraryID: tv.ID, RelativePath: path, Title: name.Title}, name, 1); err != nil {
		t.Fatal(err)
	}
	key := agent.NewKey(user.ID, agent.CreateKey{Name: "Test Agent", Access: "read_write", ExpiresInDays: 90}, "fingerprint")
	key, err = db.CreateAgentKey(ctx, key, auth.TokenHash("test-agent-secret"))
	if err != nil {
		t.Fatal(err)
	}
	return db, key, tv.ID
}

func TestAgentMetadataVersionAndAutomaticOverwriteProtection(t *testing.T) {
	for _, kind := range []string{"movie", "tv"} {
		t.Run(kind, func(t *testing.T) {
			db, key, tvLibrary := agentStoreFixture(t)
			ctx := context.Background()
			service := agent.Service{Repository: db}
			target, err := service.Target(ctx, key, kind, 1, false)
			if err != nil || target.Version != 0 || len(target.Files) != 1 {
				t.Fatalf("target: %+v %v", target, err)
			}
			input := agent.Input{Title: "Agent 中文名", OriginalLanguage: "ja", Aliases: []string{"Alias A", "Alias B"}, IDs: map[string]string{"imdb": "tt123"}}
			version, err := service.Update(ctx, key, kind, 1, 0, input, "agent")
			if err != nil || version != 1 {
				t.Fatalf("update: %d %v", version, err)
			}
			target, err = service.Target(ctx, key, kind, 1, false)
			if err != nil || !target.Metadata.Matched() || target.Metadata.TMDBID != 0 || target.Metadata.BangumiID != 0 || target.Version != 1 {
				t.Fatalf("ID-free metadata: %+v %v", target, err)
			}
			if target.Metadata.ExternalIDs["imdb"] != "tt123" || target.Metadata.Aliases != "Alias A\nAlias B" {
				t.Fatal("external IDs lost")
			}
			if _, err := service.Update(ctx, key, kind, 1, 0, agent.Input{Title: "stale"}, "agent"); !errors.Is(err, metadata.ErrConflict) {
				t.Fatal("stale version accepted")
			}
			var repo interface {
				SaveMovieMetadata(context.Context, int64, metadata.Movie) error
				RecordScrapeFailure(context.Context, int64, string, string) error
				RecordScrapeReview(context.Context, int64, []metadata.Candidate) error
				ScrapeTargets(context.Context, metadata.ScrapeRequest) ([]media.Item, error)
			} = db
			if kind == "tv" {
				repo = TelevisionRepository{Store: db}
			}
			if err := repo.SaveMovieMetadata(ctx, 1, metadata.Movie{TMDBID: 100, Title: "in-flight automatic result", Status: "matched"}); !errors.Is(err, metadata.ErrConflict) {
				t.Fatal("automatic scrape overwrote agent confirmation")
			}
			if err := repo.RecordScrapeFailure(ctx, 1, "failed", "provider_unavailable"); err != nil {
				t.Fatal(err)
			}
			if err := repo.RecordScrapeReview(ctx, 1, nil); err != nil {
				t.Fatal(err)
			}
			target, err = service.Target(ctx, key, kind, 1, false)
			if err != nil || target.Version != 1 || target.Metadata.Title != "Agent 中文名" || target.Metadata.Message != "" {
				t.Fatal("failure/review modified protected metadata")
			}
			targets, err := repo.ScrapeTargets(ctx, metadata.ScrapeRequest{})
			if err != nil {
				t.Fatal(err)
			}
			for _, item := range targets {
				if item.ID == 1 {
					t.Fatal("agent target was queued for scraping")
				}
			}
			scoped := key
			scoped.LibraryID = tvLibrary
			if kind == "movie" {
				if _, err := service.Target(ctx, scoped, kind, 1, false); !errors.Is(err, agent.ErrForbidden) {
					t.Fatal("cross-library target exposed")
				}
			}
			page, err := service.Tasks(ctx, key, metadata.ReportFilter{Status: "matched", Page: 1, PageSize: 20})
			if err != nil || page.Total != 1 || page.Items[0].Source != "agent" || page.Items[0].Version != 1 {
				t.Fatalf("report: %+v %v", page, err)
			}
			version, err = service.Artwork(ctx, key, target, 1, strings.Repeat("a", 64))
			if err != nil || version != 2 {
				t.Fatalf("artwork: %d %v", version, err)
			}
			target, err = service.Target(ctx, key, kind, 1, false)
			if err != nil || target.Metadata.PosterKey != strings.Repeat("a", 64) {
				t.Fatal("artwork reference lost")
			}
			var events int
			if err := db.DB.QueryRow(`SELECT count(*) FROM agent_metadata_events`).Scan(&events); err != nil || events != 2 {
				t.Fatal("missing atomic audit events")
			}
		})
	}
}

func TestConcurrentAgentWritesOnlyOneVersionWins(t *testing.T) {
	db, key, _ := agentStoreFixture(t)
	service := agent.Service{Repository: db}
	var wg sync.WaitGroup
	results := make(chan error, 2)
	for _, title := range []string{"first", "second"} {
		wg.Add(1)
		go func(title string) {
			defer wg.Done()
			_, err := service.Update(context.Background(), key, "movie", 1, 0, agent.Input{Title: title}, "agent")
			results <- err
		}(title)
	}
	wg.Wait()
	close(results)
	wins, conflicts := 0, 0
	for err := range results {
		if err == nil {
			wins++
		} else if errors.Is(err, metadata.ErrConflict) {
			conflicts++
		} else {
			t.Fatal(err)
		}
	}
	if wins != 1 || conflicts != 1 {
		t.Fatalf("wins=%d conflicts=%d", wins, conflicts)
	}
}
