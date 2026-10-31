CREATE TABLE agent_keys (
 id INTEGER PRIMARY KEY,
 user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
 library_id INTEGER REFERENCES libraries(id) ON DELETE CASCADE,
 name TEXT NOT NULL,
 token_hash TEXT NOT NULL UNIQUE,
 prefix TEXT NOT NULL,
 access TEXT NOT NULL CHECK(access IN ('read','read_write')),
 created_at INTEGER NOT NULL,
 expires_at INTEGER NOT NULL
);
ALTER TABLE movie_metadata ADD COLUMN source TEXT NOT NULL DEFAULT '';
ALTER TABLE movie_metadata ADD COLUMN source_url TEXT NOT NULL DEFAULT '';
ALTER TABLE movie_metadata ADD COLUMN external_ids_json TEXT NOT NULL DEFAULT '{}';
ALTER TABLE movie_metadata ADD COLUMN revision INTEGER NOT NULL DEFAULT 0;
ALTER TABLE tv_series ADD COLUMN metadata_version INTEGER NOT NULL DEFAULT 0;
CREATE TABLE agent_metadata_events (
 id INTEGER PRIMARY KEY,
 key_id INTEGER REFERENCES agent_keys(id) ON DELETE SET NULL,
 key_name TEXT NOT NULL,
 kind TEXT NOT NULL CHECK(kind IN ('movie','tv')),
 target_id INTEGER NOT NULL,
 revision INTEGER NOT NULL,
 source TEXT NOT NULL,
 created_at INTEGER NOT NULL
);
