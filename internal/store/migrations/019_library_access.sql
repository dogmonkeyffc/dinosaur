ALTER TABLE libraries ADD COLUMN restricted INTEGER NOT NULL DEFAULT 0 CHECK(restricted IN (0, 1));
CREATE TABLE library_grants (
 library_id INTEGER NOT NULL REFERENCES libraries(id) ON DELETE CASCADE,
 user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
 PRIMARY KEY(library_id, user_id)
);
