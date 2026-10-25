package store

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"

	"dinosaur/internal/media"
	"dinosaur/internal/metadata"
)

func TestTelevisionUpgradePreservesMovieProgressAndJobHistory(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "upgrade.db")
	old, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	old.SetMaxOpenConns(1)
	t.Cleanup(func() { old.Close() })
	if _, err := old.Exec(`CREATE TABLE schema_migrations(version TEXT PRIMARY KEY,applied_at INTEGER NOT NULL)`); err != nil {
		t.Fatal(err)
	}
	names, err := migrations.ReadDir("migrations")
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range names {
		if name.Name() >= "009" {
			break
		}
		content, err := migrations.ReadFile("migrations/" + name.Name())
		if err != nil {
			t.Fatal(err)
		}
		if _, err := old.Exec(string(content)); err != nil {
			t.Fatal(err)
		}
		if _, err := old.Exec(`INSERT INTO schema_migrations VALUES (?,0)`, name.Name()); err != nil {
			t.Fatal(err)
		}
	}
	for _, query := range []string{
		`INSERT INTO users(id,username,password_hash,role,created_at) VALUES (1,'admin','hash','admin',1)`,
		`INSERT INTO libraries(id,name,type,root_path,created_at,updated_at) VALUES (1,'Movies','movie','/example',1,1)`,
		`INSERT INTO media_items(id,library_id,relative_path,title,size,modified_ns,last_seen_scan) VALUES (1,1,'Movie.mp4','Movie',0,0,1)`,
		`INSERT INTO playback_progress VALUES (1,1,35,0,1)`,
		`INSERT INTO scrape_jobs(id,status,created_at) VALUES (1,'completed',1)`,
		`INSERT INTO scrape_job_items(job_id,source_media_id,media_id,title,filename,library_name,status,updated_at) VALUES (1,1,1,'Movie','Movie.mp4','Movies','matched',1)`,
	} {
		if _, err := old.Exec(query); err != nil {
			t.Fatal(err)
		}
	}
	if err := old.Close(); err != nil {
		t.Fatal(err)
	}
	upgraded, err := Open(ctx, path)
	if err != nil {
		t.Fatal(err)
	}
	defer upgraded.Close()
	page, err := upgraded.MediaPage(ctx, media.Filter{Page: 1, PageSize: 24})
	if err != nil || page.Total != 1 || page.Items[0].ID != 1 {
		t.Fatalf("movie index lost: %+v %v", page, err)
	}
	progress, err := upgraded.PlaybackProgress(ctx, 1, 1)
	if err != nil || progress.Position != 35 {
		t.Fatal("movie progress lost")
	}
	job, err := upgraded.ScrapeJobDetail(ctx, 1, metadata.ReportFilter{Page: 1, PageSize: 20})
	if err != nil || job.Job.MediaType != "movie" || len(job.Items) != 1 || job.Items[0].MediaID != 1 || job.Items[0].SeriesID != 0 {
		t.Fatalf("job history changed: %+v %v", job, err)
	}
}
