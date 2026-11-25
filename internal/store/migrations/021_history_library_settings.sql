CREATE TABLE history_library_settings (
 user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
 library_id INTEGER NOT NULL REFERENCES libraries(id) ON DELETE CASCADE,
 record_history INTEGER NOT NULL CHECK(record_history IN (0,1)),
 updated_at INTEGER NOT NULL,
 PRIMARY KEY(user_id,library_id)
);
