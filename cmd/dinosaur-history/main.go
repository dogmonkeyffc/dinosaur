package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"dinosaur/internal/doubanhar"
	"dinosaur/internal/history"
	"dinosaur/internal/historyassets"
	"dinosaur/internal/store"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	har := flag.String("har", "", "Douban collect HAR file")
	database := flag.String("database", "", "existing Dinosaur SQLite database")
	dataDir := flag.String("data-dir", "", "Dinosaur data directory")
	username := flag.String("user", "", "target username")
	check := flag.Bool("check", false, "validate HAR without writing database or posters")
	flag.Parse()
	if *har == "" {
		return fmt.Errorf("--har is required")
	}
	file, err := os.Open(*har)
	if err != nil {
		return err
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil {
		return err
	}
	if info.Size() > 128*1024*1024 {
		return fmt.Errorf("HAR exceeds 128 MiB")
	}
	report, err := doubanhar.Load(file)
	if err != nil {
		return err
	}
	if *check {
		return json.NewEncoder(os.Stdout).Encode(map[string]any{"records": len(report.Records), "collection_total": report.Total, "page_starts": report.Starts, "captured_posters": len(report.Posters)})
	}
	if *database == "" || *dataDir == "" || *username == "" {
		return fmt.Errorf("--database, --data-dir and --user are required; stop the service and back up the database before import")
	}
	if _, err := os.Stat(*database); err != nil {
		return fmt.Errorf("database must already exist: %w", err)
	}
	db, err := store.Open(context.Background(), *database)
	if err != nil {
		return err
	}
	defer db.Close()
	u, err := db.UserByName(context.Background(), *username)
	if err != nil {
		return err
	}
	if u.Disabled {
		return fmt.Errorf("target user is disabled")
	}
	entries := make([]history.Entry, 0, len(report.Records))
	posters := 0
	for _, record := range report.Records {
		e := record.Entry
		if data := report.Posters[record.PosterURL]; len(data) > 0 {
			e.PosterKey, err = historyassets.Put(filepath.Join(*dataDir, "history-posters"), data)
			if err != nil {
				return err
			}
			posters++
		}
		entries = append(entries, e)
	}
	result, err := db.ImportHistory(context.Background(), u.ID, entries)
	if err != nil {
		return err
	}
	return json.NewEncoder(os.Stdout).Encode(map[string]any{"user": u.Username, "imported": result.Imported, "skipped": result.Skipped, "records": len(entries), "posters": posters, "collection_total": report.Total})
}
