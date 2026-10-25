CREATE TABLE scrape_job_items (
    id INTEGER PRIMARY KEY,
    job_id INTEGER NOT NULL REFERENCES scrape_jobs(id) ON DELETE CASCADE,
    source_media_id INTEGER NOT NULL,
    media_id INTEGER REFERENCES media_items(id) ON DELETE SET NULL,
    title TEXT NOT NULL,
    filename TEXT NOT NULL,
    library_name TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'pending',
    message TEXT NOT NULL DEFAULT '',
    tmdb_id INTEGER NOT NULL DEFAULT 0,
    updated_at INTEGER NOT NULL,
    UNIQUE(job_id, source_media_id)
);
CREATE INDEX scrape_job_items_job ON scrape_job_items(job_id,id);
