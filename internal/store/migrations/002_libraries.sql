CREATE TABLE libraries (
 id INTEGER PRIMARY KEY,
 name TEXT NOT NULL,
 type TEXT NOT NULL CHECK(type IN ('movie', 'tv')),
 root_path TEXT NOT NULL UNIQUE,
 created_at INTEGER NOT NULL,
 updated_at INTEGER NOT NULL
);
