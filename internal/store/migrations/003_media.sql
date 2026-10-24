CREATE TABLE scan_jobs (
 id INTEGER PRIMARY KEY AUTOINCREMENT,
 library_id INTEGER NOT NULL REFERENCES libraries(id) ON DELETE CASCADE,
 status TEXT NOT NULL CHECK(status IN ('queued','scanning','completed','failed','interrupted')),
 discovered INTEGER NOT NULL DEFAULT 0,
 processed INTEGER NOT NULL DEFAULT 0,
 errors INTEGER NOT NULL DEFAULT 0,
 message TEXT NOT NULL DEFAULT '',
 created_at INTEGER NOT NULL,
 finished_at INTEGER NOT NULL DEFAULT 0
);
CREATE INDEX scan_jobs_library ON scan_jobs(library_id,id DESC);
CREATE UNIQUE INDEX scan_jobs_active ON scan_jobs(library_id) WHERE status IN ('queued','scanning');
CREATE TABLE media_items (
 id INTEGER PRIMARY KEY AUTOINCREMENT,
 library_id INTEGER NOT NULL REFERENCES libraries(id) ON DELETE CASCADE,
 relative_path TEXT NOT NULL,
 title TEXT NOT NULL,
 year INTEGER NOT NULL DEFAULT 0,
 size INTEGER NOT NULL,
 modified_ns INTEGER NOT NULL,
 duration REAL NOT NULL DEFAULT 0,
 format TEXT NOT NULL DEFAULT '',
 video_codec TEXT NOT NULL DEFAULT '',
 width INTEGER NOT NULL DEFAULT 0,
 height INTEGER NOT NULL DEFAULT 0,
 tracks_json TEXT NOT NULL DEFAULT '[]',
 poster_key TEXT NOT NULL DEFAULT '',
 probe_error INTEGER NOT NULL DEFAULT 0,
 missing INTEGER NOT NULL DEFAULT 0,
 last_seen_scan INTEGER NOT NULL,
 UNIQUE(library_id,relative_path)
);
CREATE INDEX media_library_title ON media_items(library_id,title);
