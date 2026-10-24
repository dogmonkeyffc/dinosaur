CREATE TABLE playback_progress (
 user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
 media_id INTEGER NOT NULL REFERENCES media_items(id) ON DELETE CASCADE,
 position REAL NOT NULL DEFAULT 0 CHECK(position>=0),
 watched INTEGER NOT NULL DEFAULT 0 CHECK(watched IN (0,1)),
 updated_at INTEGER NOT NULL,
 PRIMARY KEY(user_id,media_id)
);
CREATE INDEX playback_recent ON playback_progress(user_id,updated_at DESC);
