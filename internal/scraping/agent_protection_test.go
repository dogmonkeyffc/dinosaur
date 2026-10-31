package scraping

import (
	"context"
	"testing"
	"time"

	"dinosaur/internal/agent"
	"dinosaur/internal/auth"
	"dinosaur/internal/metadata"
)

func TestAgentConfirmationSurvivesAlreadyRunningAutomaticLookup(t *testing.T) {
	manager, db, provider := scrapeFixture(t)
	provider.block = make(chan struct{})
	job, err := manager.Start(context.Background(), metadata.ScrapeRequest{})
	if err != nil {
		t.Fatal(err)
	}
	deadline := time.Now().Add(time.Second)
	for {
		searches, _ := provider.calls()
		if searches > 0 {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("automatic lookup did not start")
		}
		time.Sleep(time.Millisecond)
	}
	user, err := db.CreateUser(context.Background(), "curator", "hash", "admin", true)
	if err != nil {
		t.Fatal(err)
	}
	key := agent.NewKey(user.ID, agent.CreateKey{Name: "Curator", Access: "read_write", ExpiresInDays: 90}, "test")
	key, err = db.CreateAgentKey(context.Background(), key, auth.TokenHash("test-curator-key"))
	if err != nil {
		t.Fatal(err)
	}
	service := agent.Service{Repository: db}
	if _, err := service.Update(context.Background(), key, "movie", 1, 0, agent.Input{Title: "Agent 已确认"}, "agent"); err != nil {
		t.Fatal(err)
	}
	close(provider.block)
	waitScrape(t, manager, db, job.ID)
	item, err := db.MediaByID(context.Background(), 1)
	if err != nil || item.Title != "Agent 已确认" || item.Metadata.Source != "agent" || item.Metadata.Version != 1 || item.Metadata.Message != "" {
		t.Fatal("in-flight result overwrote or invalidated the Agent confirmation")
	}
}
