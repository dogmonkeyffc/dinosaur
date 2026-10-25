package store

import (
	"context"
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"

	"dinosaur/internal/metadata"
)

func TestMetadataSettingsPersistAcrossReopen(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "settings.db")
	db, err := Open(ctx, path)
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.UpdateMetadataSettings(ctx, metadata.SettingsUpdate{
		CredentialType: "api_key", Language: "zh-CN", Credential: "persistent-private-key",
	}, false)
	if err != nil {
		db.Close()
		t.Fatal(err)
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	db, err = Open(ctx, path)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	saved, err := db.MetadataSettings(ctx)
	if err != nil || saved.Credential != "persistent-private-key" || saved.CredentialType != "api_key" {
		t.Fatal("configuration did not persist")
	}
	data, err := json.Marshal(saved)
	if err != nil || strings.Contains(string(data), saved.Credential) {
		t.Fatal("settings JSON exposes secret")
	}
}

func TestScrapeRecoveryInterruptsOnlyActiveJobs(t *testing.T) {
	ctx := context.Background()
	db, err := Open(ctx, filepath.Join(t.TempDir(), "jobs.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	completed, err := db.NewScrapeJob(ctx, metadata.ScrapeRequest{})
	if err != nil {
		t.Fatal(err)
	}
	completed.Status, completed.Matched = "completed", 2
	if err := db.SaveScrapeJob(ctx, completed); err != nil {
		t.Fatal(err)
	}
	if _, err := db.NewScrapeJob(ctx, metadata.ScrapeRequest{}); err != nil {
		t.Fatal(err)
	}
	if err := db.InterruptScrapes(ctx); err != nil {
		t.Fatal(err)
	}
	jobs, err := db.ScrapeJobs(ctx)
	if err != nil || len(jobs) != 2 || jobs[0].Status != "interrupted" || jobs[1].Matched != 2 || jobs[1].Status != "completed" {
		t.Fatal("restart recovery lost job history")
	}
}
