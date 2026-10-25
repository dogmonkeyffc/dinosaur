CREATE TABLE movie_metadata (
    media_id INTEGER PRIMARY KEY REFERENCES media_items(id) ON DELETE CASCADE,
    tmdb_id INTEGER NOT NULL DEFAULT 0,
    title TEXT NOT NULL DEFAULT '',
    original_title TEXT NOT NULL DEFAULT '',
    overview TEXT NOT NULL DEFAULT '',
    year INTEGER NOT NULL DEFAULT 0,
    release_date TEXT NOT NULL DEFAULT '',
    language TEXT NOT NULL DEFAULT '',
    poster_key TEXT NOT NULL DEFAULT '',
    aliases TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL,
    message TEXT NOT NULL DEFAULT '',
    updated_at INTEGER NOT NULL
);
CREATE TABLE scrape_jobs (
    id INTEGER PRIMARY KEY,
    library_id INTEGER NOT NULL DEFAULT 0,
    media_id INTEGER NOT NULL DEFAULT 0,
    status TEXT NOT NULL,
    total INTEGER NOT NULL DEFAULT 0,
    processed INTEGER NOT NULL DEFAULT 0,
    matched INTEGER NOT NULL DEFAULT 0,
    review INTEGER NOT NULL DEFAULT 0,
    errors INTEGER NOT NULL DEFAULT 0,
    skipped INTEGER NOT NULL DEFAULT 0,
    message TEXT NOT NULL DEFAULT '',
    created_at INTEGER NOT NULL,
    ended_at INTEGER NOT NULL DEFAULT 0
);
